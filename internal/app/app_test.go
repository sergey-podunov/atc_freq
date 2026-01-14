package app_test

import (
	"atc_freq/internal/app"
	"atc_freq/internal/sim"
	"atc_freq/internal/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApp_GetAirportFrequencies(t *testing.T) {
	tests := []struct {
		name          string
		icao          string
		mockSetup     func(*sim.MockConnection)
		expectedFreqs []sim.AirportFrequency
	}{
		{
			name: "successful retrieval of frequencies",
			icao: "KJFK",
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
		},
		{
			name: "empty icao",
			icao: "",
			mockSetup: func(m *sim.MockConnection) {
				// No expectations - should fail before any calls
			},
			expectedFreqs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := new(sim.MockConnection)
			tt.mockSetup(mockConn)

			application := app.NewApp(mockConn)

			freqs, err := application.GetFrequencies(tt.icao)

			if tt.expectedFreqs == nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFreqs, freqs)
			}

			mockConn.AssertExpectations(t)
		})
	}
}
