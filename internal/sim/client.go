package sim

import (
	"atc_freq/internal/helpers"
	"fmt"
	"math"
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

// CloudDensity represents interpreted cloud density at a grid point
type CloudDensity struct {
	Value      byte    // Raw density value (0-255)
	Percentage float64 // Density as percentage (0-100)
	Coverage   string  // Human-readable coverage level
	MinAlt     int     // Minimum altitude in feet
	MaxAlt     int     // Maximum altitude in feet
}

// Coordinates represents geographic coordinates
type Coordinates struct {
	Lat float64
	Lon float64
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
			exception := (*SIMCONNECT_RECV_EXCEPTION)(unsafe.Pointer(ppData))
			return nil, fmt.Errorf("connection exception received: %d (sendID: %d, index: %d)", exception.DwException, exception.DwSendID, exception.DwIndex)
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
	WEATHER_REQUEST_ID_BASE     = 0x3001
	CLOUD_STATE_REQUEST_ID_BASE = 0x4001
	WAYPOINT_DEFINE_ID          = 0x1002
	WAYPOINT_REQUEST_ID         = 0x2002
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
			return nil, fmt.Errorf("connection exception received, ID: %d", ppData.DwID)
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

// GetWaypointCoordinates retrieves coordinates for the specified waypoints/airports
func (client *Client) GetWaypointCoordinates(waypoints []string, timeout time.Duration) (map[string]Coordinates, error) {
	if len(waypoints) == 0 {
		return nil, fmt.Errorf("no waypoints provided")
	}

	connection := client.simConnection
	err := connection.Open("go-coords-client")
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	// Define facility structure to get lat/lon
	if err := connection.AddField("OPEN AIRPORT", WAYPOINT_DEFINE_ID); err != nil {
		return nil, err
	}
	if err := connection.AddField("LATITUDE", WAYPOINT_DEFINE_ID); err != nil {
		return nil, err
	}
	if err := connection.AddField("LONGITUDE", WAYPOINT_DEFINE_ID); err != nil {
		return nil, err
	}
	if err := connection.AddField("CLOSE AIRPORT", WAYPOINT_DEFINE_ID); err != nil {
		return nil, err
	}

	// Request facility data for each waypoint to get coordinates
	requestIDToWaypoint := make(map[uint32]string)
	for i, wp := range waypoints {
		wp = strings.ToUpper(strings.TrimSpace(wp))
		if wp == "" {
			continue
		}
		requestID := uint32(WAYPOINT_REQUEST_ID + i)
		requestIDToWaypoint[requestID] = wp

		err = connection.RequestFacilityData(wp, "", WAYPOINT_DEFINE_ID, requestID)
		if err != nil {
			return nil, fmt.Errorf("failed to request facility data for %s: %w", wp, err)
		}
	}

	result := make(map[string]Coordinates)
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
			return nil, fmt.Errorf("connection exception during facility request")
		case SIMCONNECT_RECV_ID_FACILITY_DATA:
			facData := (*SIMCONNECT_RECV_FACILITY_DATA)(unsafe.Pointer(ppData))
			wp, exists := requestIDToWaypoint[facData.UserRequestId]
			if !exists {
				continue
			}

			// Extract lat/lon from facility data
			// When using AddField with LATITUDE/LONGITUDE, data contains only requested fields
			dataPtr := unsafe.Pointer(&facData.Data)
			latPtr := (*float64)(dataPtr)                              // LATITUDE at offset 0
			lonPtr := (*float64)(unsafe.Pointer(uintptr(dataPtr) + 8)) // LONGITUDE at offset 8

			result[wp] = Coordinates{
				Lat: *latPtr,
				Lon: *lonPtr,
			}

		case SIMCONNECT_RECV_ID_FACILITY_DATA_END:
			end := (*SIMCONNECT_RECV_FACILITY_DATA_END)(unsafe.Pointer(ppData))
			if _, exists := requestIDToWaypoint[end.RequestId]; exists {
				pendingRequests--
			}
		}
	}

	if pendingRequests > 0 {
		return nil, fmt.Errorf("timeout getting waypoint coordinates: %d/%d received", len(result), len(requestIDToWaypoint))
	}

	return result, nil
}

// GetCloudDensityByCoordinates retrieves cloud density at the center of a grid for specified coordinates and altitude range
func (client *Client) GetCloudDensityByCoordinates(coords Coordinates, minAlt, maxAlt float32, timeout time.Duration) (CloudDensity, error) {
	connection := client.simConnection
	err := connection.Open("go-cloud-density-client")
	if err != nil {
		return CloudDensity{}, err
	}
	defer connection.Close()

	requestID := uint32(CLOUD_STATE_REQUEST_ID_BASE)

	// Calculate offsets for 5x5 km box (±2.5 km from center)
	// 1 degree of latitude ≈ 111 km
	// 1 degree of longitude ≈ 111 km * cos(latitude)
	const kmPerDegreeLat = 111.0
	const halfBoxKm = 2.5

	latOffset := halfBoxKm / kmPerDegreeLat
	lonOffset := halfBoxKm / (kmPerDegreeLat * math.Cos(coords.Lat*math.Pi/180))

	err = connection.RequestCloudState(
		requestID,
		float32(coords.Lat-latOffset), float32(coords.Lon-lonOffset), minAlt,
		float32(coords.Lat+latOffset), float32(coords.Lon+lonOffset), maxAlt,
	)
	if err != nil {
		return CloudDensity{}, fmt.Errorf("failed to request cloud state: %w", err)
	}

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		ppData, ok := connection.GetNextDispatch()
		if !ok {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		switch ppData.DwID {
		case SIMCONNECT_RECV_ID_EXCEPTION:
			return CloudDensity{}, fmt.Errorf("connection exception during cloud state request")
		case SIMCONNECT_RECV_ID_CLOUD_STATE:
			cloudData := (*SIMCONNECT_RECV_CLOUD_STATE)(unsafe.Pointer(ppData))
			if cloudData.DwRequestID != requestID {
				continue
			}

			// Extract raw cloud data
			//dataPtr := unsafe.Pointer(uintptr(unsafe.Pointer(cloudData)) + unsafe.Sizeof(*cloudData))
			dataPtr := unsafe.Pointer(&cloudData.RgbData[0])
			rawData := make([]byte, cloudData.DwArraySize)
			for i := uint32(0); i < cloudData.DwArraySize; i++ {
				rawData[i] = *(*byte)(unsafe.Pointer(uintptr(dataPtr) + uintptr(i)))
			}

			fmt.Printf("raw cloud data between %f0.1 - %f0.1", minAlt, maxAlt)
			for _, val := range rawData {
				density := interpretCloudDensity(val)
				fmt.Printf(" %s - %.2f%%, ", density.Coverage, density.Percentage)
			}

			// Return density at center of 64x64 grid (position 32,32)
			const gridSize = 64
			centerIndex := (gridSize/2)*gridSize + (gridSize / 2)
			if centerIndex < len(rawData) {
				return interpretCloudDensity(rawData[centerIndex]), nil
			}

			return CloudDensity{}, nil
		}
	}

	return CloudDensity{}, fmt.Errorf("timeout waiting for cloud state response")
}

// interpretCloudDensity converts a raw density byte into a CloudDensity struct
func interpretCloudDensity(value byte) CloudDensity {
	percentage := (float64(value) / 255.0) * 100.0

	var coverage string
	switch {
	case value == 0:
		coverage = "CLR" // Clear
	case value < 64:
		coverage = "FEW" // Few (1/8 to 2/8)
	case value < 128:
		coverage = "SCT" // Scattered (3/8 to 4/8)
	case value < 192:
		coverage = "BKN" // Broken (5/8 to 7/8)
	default:
		coverage = "OVC" // Overcast (8/8)
	}

	return CloudDensity{
		Value:      value,
		Percentage: percentage,
		Coverage:   coverage,
	}
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
