package sim

import (
	"fmt"
	"strings"
	"time"
)

const clientTimeout = 10 * time.Second
const altStep = 500
const maxAltitude = 10000

// Service is a layer between applications and Client
type Service struct {
	client *Client
}

// NewService creates a new Service with the provided Client
func NewService(client *Client) *Service {
	return &Service{
		client: client,
	}
}

// GetFrequency retrieves all frequencies for the specified ICAO airport code
func (s *Service) GetFrequency(icao string) ([]AirportFrequency, error) {
	icao = strings.ToUpper(strings.TrimSpace(icao))
	if icao == "" {
		return nil, fmt.Errorf("ICAO code cannot be empty")
	}

	freqs, err := s.client.GetAirportFrequencies(icao, clientTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to get frequencies for %s: %w", icao, err)
	}

	return freqs, nil
}

// GetWeather retrieves weather information for the specified waypoints
func (s *Service) GetWeather(waypoints []string) (map[string]*Weather, error) {
	if len(waypoints) == 0 {
		return nil, fmt.Errorf("no waypoints provided")
	}

	cleanedWaypoints := cleanWaypoints(waypoints)
	if len(cleanedWaypoints) == 0 {
		return nil, fmt.Errorf("no valid waypoints provided")
	}

	return s.client.GetWeather(cleanedWaypoints, clientTimeout)
}

// GetCloudDensity retrieves cloud density at multiple altitude layers for each waypoint
func (s *Service) GetCloudDensity(waypoints []string) (map[string][]CloudDensity, error) {
	if len(waypoints) == 0 {
		return nil, fmt.Errorf("no waypoints provided")
	}

	cleanedWaypoints := cleanWaypoints(waypoints)
	if len(cleanedWaypoints) == 0 {
		return nil, fmt.Errorf("no valid waypoints provided")
	}

	// Get coordinates for all waypoints
	fmt.Printf("Getting coordinates for %d waypoints...\n", len(cleanedWaypoints))
	coords, err := s.client.GetWaypointCoordinates(cleanedWaypoints, clientTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to get waypoint coordinates: %w", err)
	}
	fmt.Printf("Coordinates retrieved: %s\n", coords)

	fmt.Println("Getting ambient cloud state...")
	inCloud, err := s.client.GetAmbientInCloud(clientTimeout * 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get ambient in cloud: %w", err)
	}

	// Get cloud density for each waypoint at all altitude layers (0-10000 feet, 500 feet step)
	result := make(map[string][]CloudDensity)
	var cloudDensity CloudDensity
	if inCloud {
		cloudDensity = CloudDensity{
			Coverage:   "OVERCAST",
			Value:      255,
			MinAlt:     3000,
			MaxAlt:     3500,
			Percentage: 100,
		}
	} else {
		cloudDensity = CloudDensity{
			Coverage:   "CLR",
			Value:      0,
			MinAlt:     3000,
			MaxAlt:     3500,
			Percentage: 0,
		}
	}

	result[cleanedWaypoints[0]] = append(result[cleanedWaypoints[0]], cloudDensity)

	/*	for wp, coord := range coords {
		var layers []CloudDensity
		for minAlt := 0; minAlt < maxAltitude; minAlt += altStep {
			maxAlt := minAlt + altStep
			density, err := s.client.GetCloudDensityByCoordinates(coord, float32(minAlt), float32(maxAlt), clientTimeout)
			if err != nil {
				return nil, fmt.Errorf("failed to get cloud density for %s at %d-%d ft: %w", wp, minAlt, maxAlt, err)
			}
			density.MinAlt = minAlt
			density.MaxAlt = maxAlt
			layers = append(layers, density)
		}
		result[wp] = layers
	}*/

	return result, nil
}

func cleanWaypoints(waypoints []string) []string {
	cleaned := make([]string, 0, len(waypoints))
	for _, wp := range waypoints {
		wp = strings.TrimSpace(wp)
		if wp != "" {
			cleaned = append(cleaned, wp)
		}
	}
	return cleaned
}
