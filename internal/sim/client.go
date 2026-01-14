package sim

import (
	"atc_freq/internal/helpers"
	"fmt"
	"strings"
	"time"
	"unsafe"
)

type FACILITY_FREQUENCY_DATA struct {
	TYPE      int32
	FREQUENCY int32    // Hz
	NAME      [64]byte // C char[64]
}

type AirportFrequency struct {
	Type     string
	TypeCode int32
	Name     string
	Hz       int
	MHz      float64
}

// CloudLayer represents a single cloud layer with base altitude and coverage
type CloudLayer struct {
	Base     int    // Base altitude in feet
	Coverage string // Cloud coverage (FEW, SCT, BKN, OVC)
}

// Weather contains weather information for a waypoint
type Weather struct {
	Waypoint   string
	Visibility int // Visibility in statute miles (0-10+)
	Clouds     []CloudLayer
	RawMetar   string // Raw METAR string from sim
}

var freqTypeMap = map[int32]string{
	0:  "NONE",
	1:  "ATIS",
	2:  "MULTICOM",
	3:  "UNICOM",
	4:  "CTAF",
	5:  "GROUND",
	6:  "TOWER",
	7:  "CLEARANCE",
	8:  "APPROACH",
	9:  "DEPARTURE",
	10: "CENTER",
	11: "FSS",
	12: "AWOS",
	13: "ASOS",
	14: "CPT",
	15: "GCO",
}

type Client struct {
	simConnection Connection
}

func NewClient(conn Connection) *Client {
	return &Client{simConnection: conn}
}

func (client *Client) GetAirportFrequencies(icao string, timeout time.Duration) ([]AirportFrequency, error) {
	icao = strings.ToUpper(strings.TrimSpace(icao))
	if icao == "" {
		return nil, fmt.Errorf("icao is empty")
	}

	connection := client.simConnection
	err := connection.Open("go-freq-client")
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	// Facility definition: OPEN AIRPORT -> OPEN FREQUENCY -> TYPE/FREQUENCY/NAME -> CLOSE -> CLOSE
	if err := connection.AddField("OPEN AIRPORT", DEFINE_ID); err != nil {
		return nil, err
	}
	if err := connection.AddField("OPEN FREQUENCY", DEFINE_ID); err != nil {
		return nil, err
	}
	if err := connection.AddField("TYPE", DEFINE_ID); err != nil {
		return nil, err
	}
	if err := connection.AddField("FREQUENCY", DEFINE_ID); err != nil {
		return nil, err
	}
	if err := connection.AddField("NAME", DEFINE_ID); err != nil {
		return nil, err
	}
	if err := connection.AddField("CLOSE FREQUENCY", DEFINE_ID); err != nil {
		return nil, err
	}
	if err := connection.AddField("CLOSE AIRPORT", DEFINE_ID); err != nil {
		return nil, err
	}

	err = connection.RequestFacilityData(icao, "", DEFINE_ID, REQUEST_ID)
	if err != nil {
		return nil, err
	}

	var out []AirportFrequency
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		ppData, ok := connection.GetNextDispatch()
		if !ok {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		fmt.Printf("dispatch: %#v\n", ppData)

		switch ppData.DwID {
		case SIMCONNECT_RECV_ID_EXCEPTION:
			// Cast to SIMCONNECT_RECV_EXCEPTION and print dwException
			fmt.Printf("Connection Exception received! ID: %d\n", ppData.DwID)
		case SIMCONNECT_RECV_ID_FACILITY_DATA:
			facData := (*SIMCONNECT_RECV_FACILITY_DATA)(unsafe.Pointer(ppData))

			if facData.UserRequestId != REQUEST_ID {
				continue
			}

			if facData.Type != SIMCONNECT_FACILITY_DATA_FREQUENCY {
				continue
			}

			fmt.Println("Found Frequency Data!")

			// Frequency data starts at the Data field
			dataPtr := unsafe.Pointer(&facData.Data)
			freq := (*FACILITY_FREQUENCY_DATA)(dataPtr)

			name := helpers.TrimCString(freq.NAME[:])
			hz := int(freq.FREQUENCY)
			tcode := freq.TYPE
			tname := freqTypeMap[tcode]
			if tname == "" {
				tname = fmt.Sprintf("UNKNOWN_%d", tcode)
			}

			out = append(out, AirportFrequency{
				Type:     tname,
				TypeCode: tcode,
				Name:     name,
				Hz:       hz,
				MHz:      helpers.HzToMHz(hz),
			})

		case SIMCONNECT_RECV_ID_FACILITY_DATA_END:
			fmt.Println("got facility data end")
			end := (*SIMCONNECT_RECV_FACILITY_DATA_END)(unsafe.Pointer(ppData))
			if end.RequestId == REQUEST_ID {
				return out, nil
			}
		}
	}

	return out, fmt.Errorf("timeout waiting for facility data end (got %d frequencies so far)", len(out))
}

const (
	WEATHER_REQUEST_ID_BASE = 0x3001
)

// GetWeather retrieves weather information for the specified waypoints
func (client *Client) GetWeather(waypoints []string, timeout time.Duration) (map[string]*Weather, error) {
	if len(waypoints) == 0 {
		return nil, fmt.Errorf("no waypoints provided")
	}

	connection := client.simConnection
	err := connection.Open("go-weather-client")
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	// Request weather for each waypoint
	requestIDToWaypoint := make(map[uint32]string)
	for i, wp := range waypoints {
		wp = strings.ToUpper(strings.TrimSpace(wp))
		if wp == "" {
			continue
		}
		requestID := uint32(WEATHER_REQUEST_ID_BASE + i)
		requestIDToWaypoint[requestID] = wp

		err = connection.RequestWeatherObservation(wp, requestID)
		if err != nil {
			return nil, fmt.Errorf("failed to request weather for %s: %w", wp, err)
		}
	}

	result := make(map[string]*Weather)
	pendingRequests := len(requestIDToWaypoint)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) && pendingRequests > 0 {
		ppData, ok := connection.GetNextDispatch()
		if !ok {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		switch ppData.DwID {
		case SIMCONNECT_RECV_ID_EXCEPTION:
			fmt.Printf("Weather Exception received! ID: %d\n", ppData.DwID)
		case SIMCONNECT_RECV_ID_WEATHER_OBSERVATION:
			weatherData := (*SIMCONNECT_RECV_WEATHER_OBSERVATION)(unsafe.Pointer(ppData))

			wp, exists := requestIDToWaypoint[weatherData.DwRequestID]
			if !exists {
				continue
			}

			// Extract METAR string - starts at SzMetar field and goes to end of struct
			metarPtr := unsafe.Pointer(&weatherData.SzMetar[0])
			metar := helpers.TrimCString((*[512]byte)(metarPtr)[:])

			weather := parseMetar(wp, metar)
			result[wp] = weather
			pendingRequests--
		}
	}

	if pendingRequests > 0 {
		return result, fmt.Errorf("timeout: received %d of %d weather responses", len(result), len(requestIDToWaypoint))
	}

	return result, nil
}

// parseMetar parses a METAR string and extracts visibility and cloud layers
func parseMetar(waypoint, metar string) *Weather {
	weather := &Weather{
		Waypoint: waypoint,
		RawMetar: metar,
		Clouds:   []CloudLayer{},
	}

	parts := strings.Fields(metar)
	for i, part := range parts {
		// Parse visibility (format: XXXXSM where XXXX is visibility in SM)
		if strings.HasSuffix(part, "SM") {
			visStr := strings.TrimSuffix(part, "SM")
			// Handle fractional visibility like "1/2SM" or "1 1/2SM"
			if strings.Contains(visStr, "/") {
				// Fractional - treat as less than 1
				weather.Visibility = 0
			} else {
				var vis int
				fmt.Sscanf(visStr, "%d", &vis)
				weather.Visibility = vis
			}
		}

		// Parse cloud layers (format: XXXnnn where XXX is coverage, nnn is altitude in hundreds of feet)
		if len(part) >= 6 {
			coverage := part[:3]
			if coverage == "FEW" || coverage == "SCT" || coverage == "BKN" || coverage == "OVC" {
				altStr := part[3:]
				// Remove any suffix like CB (cumulonimbus)
				altStr = strings.TrimRight(altStr, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
				var alt int
				if _, err := fmt.Sscanf(altStr, "%d", &alt); err == nil {
					weather.Clouds = append(weather.Clouds, CloudLayer{
						Base:     alt * 100, // Convert from hundreds to feet
						Coverage: coverage,
					})
				}
			}
		}

		// Also check for CLR or SKC (clear sky)
		if part == "CLR" || part == "SKC" {
			// No clouds
		}

		// Check for CAVOK (visibility > 10km, no clouds below 5000ft)
		if part == "CAVOK" {
			weather.Visibility = 10
		}

		// Handle P6SM (visibility greater than 6 SM - used in US METAR)
		if part == "P6SM" {
			weather.Visibility = 10
		}

		// Handle visibility in meters (4 digits, typically European format)
		if len(part) == 4 && i > 0 {
			var visMeters int
			if _, err := fmt.Sscanf(part, "%d", &visMeters); err == nil && visMeters >= 0 && visMeters <= 9999 {
				// Convert meters to statute miles (roughly)
				weather.Visibility = visMeters / 1609
				if weather.Visibility > 10 {
					weather.Visibility = 10
				}
			}
		}
	}

	return weather
}
