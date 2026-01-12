package sim_test

import (
	"atc_freq/internal/sim"
	"atc_freq/internal/testutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetAirportFrequencies(t *testing.T) {
	tests := []struct {
		name          string
		icao          string
		timeout       time.Duration
		mockSetup     func(*sim.MockConnection)
		expectedFreqs []sim.AirportFrequency
		expectedError string
	}{
		{
			name:    "successful retrieval of frequencies",
			icao:    "KJFK",
			timeout: 5000 * time.Second,
			mockSetup: func(m *sim.MockConnection) {
				// Expect Open call
				m.On("Open", "go-freq-client").Return(nil).Once()

				// Expect all addField calls
				m.On("addField", "OPEN AIRPORT", uint32(sim.DEFINE_ID)).Return(nil).Once()
				m.On("addField", "OPEN FREQUENCY", uint32(sim.DEFINE_ID)).Return(nil).Once()
				m.On("addField", "TYPE", uint32(sim.DEFINE_ID)).Return(nil).Once()
				m.On("addField", "FREQUENCY", uint32(sim.DEFINE_ID)).Return(nil).Once()
				m.On("addField", "NAME", uint32(sim.DEFINE_ID)).Return(nil).Once()
				m.On("addField", "CLOSE FREQUENCY", uint32(sim.DEFINE_ID)).Return(nil).Once()
				m.On("addField", "CLOSE AIRPORT", uint32(sim.DEFINE_ID)).Return(nil).Once()

				// Expect RequestFacilityData
				m.On("RequestFacilityData", "KJFK", "", uint32(sim.DEFINE_ID), uint32(sim.REQUEST_ID)).Return(nil).Once()

				// Mock frequency data response
				freqData := testutil.CreateFacilityDataResponse(6, 118700000, "Tower")
				m.On("GetNextDispatch").Return(freqData, true).Once()

				// Mock facility data end
				endData := testutil.CreateFacilityDataEndResponse()
				m.On("GetNextDispatch").Return(endData, true).Once()

				// Expect Close call
				m.On("Close").Return().Once()
			},
			expectedFreqs: []sim.AirportFrequency{
				{
					Type:     "TOWER",
					TypeCode: 6,
					Name:     "Tower",
					Hz:       118700000,
					MHz:      118.7,
				},
			},
			expectedError: "",
		},
		{
			name:    "empty icao",
			icao:    "",
			timeout: 5 * time.Second,
			mockSetup: func(m *sim.MockConnection) {
				// No expectations - should fail before any calls
			},
			expectedFreqs: nil,
			expectedError: "icao is empty",
		},
		{
			name:    "timeout waiting for data",
			icao:    "KLAX",
			timeout: 100 * time.Millisecond,
			mockSetup: func(m *sim.MockConnection) {
				m.On("Open", "go-freq-client").Return(nil).Once()
				m.On("addField", mock.Anything, mock.Anything).Return(nil).Times(7)
				m.On("RequestFacilityData", "KLAX", "", uint32(sim.DEFINE_ID), uint32(sim.REQUEST_ID)).Return(nil).Once()

				// Never return valid data, causing timeout
				m.On("GetNextDispatch").Return(nil, false).Maybe()

				m.On("Close").Return().Once()
			},
			expectedFreqs: []sim.AirportFrequency{},
			expectedError: "timeout waiting for facility data end (got 0 frequencies so far)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := new(sim.MockConnection)
			tt.mockSetup(mockConn)

			client := sim.NewClientWithConnection(mockConn)

			freqs, err := client.GetAirportFrequencies(tt.icao, tt.timeout)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFreqs, freqs)
			}

			mockConn.AssertExpectations(t)
		})
	}
}