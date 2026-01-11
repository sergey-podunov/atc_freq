package sim

import (
	"fmt"
	"strings"
	"time"
)

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

	freqs, err := s.client.GetAirportFrequencies(icao, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to get frequencies for %s: %w", icao, err)
	}

	return freqs, nil
}

// GetWeather retrieves weather information for the specified waypoints
func (s *Service) GetWeather(waypoints []string) (map[string]string, error) {
	if len(waypoints) == 0 {
		return nil, fmt.Errorf("no waypoints provided")
	}

	// Clean up waypoints
	cleanedWaypoints := make([]string, 0, len(waypoints))
	for _, wp := range waypoints {
		wp = strings.ToUpper(strings.TrimSpace(wp))
		if wp != "" {
			cleanedWaypoints = append(cleanedWaypoints, wp)
		}
	}

	if len(cleanedWaypoints) == 0 {
		return nil, fmt.Errorf("no valid waypoints provided")
	}

	// TODO: Implement actual weather retrieval logic
	// For now, return placeholder data
	result := make(map[string]string)
	for _, wp := range cleanedWaypoints {
		result[wp] = "(weather data not yet implemented)"
	}

	return result, nil
}
