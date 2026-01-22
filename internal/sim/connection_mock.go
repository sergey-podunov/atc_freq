package sim

import (
	"github.com/stretchr/testify/mock"
)

type MockConnection struct {
	mock.Mock
}

func (m *MockConnection) Open(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockConnection) Close() {
	m.Called()
}

func (m *MockConnection) AddField(field string, defineID uint32) error {
	args := m.Called(field, defineID)
	return args.Error(0)
}

func (m *MockConnection) AddToDataDefinition(defineID uint32, datumName string, unitsName string, datumType uint32) error {
	args := m.Called(defineID, datumName, unitsName, datumType)
	return args.Error(0)
}

func (m *MockConnection) RequestFacilityData(icao string, region string, defineID uint32, requestID uint32) error {
	args := m.Called(icao, region, defineID, requestID)
	return args.Error(0)
}

func (m *MockConnection) RequestWeatherObservation(icao string, requestID uint32) error {
	args := m.Called(icao, requestID)
	return args.Error(0)
}

func (m *MockConnection) RequestCloudState(requestID uint32, minLat, minLon, minAlt, maxLat, maxLon, maxAlt float32) error {
	args := m.Called(requestID, minLat, minLon, minAlt, maxLat, maxLon, maxAlt)
	return args.Error(0)
}

func (m *MockConnection) RequestDataOnSimObjectType(requestID, defineID uint32, radius uint32, objectType uint32) error {
	args := m.Called(requestID, defineID, radius, objectType)
	return args.Error(0)
}

func (m *MockConnection) RequestDataOnSimObject(requestID, defineID, objectID, period uint32) error {
	args := m.Called(requestID, defineID, objectID, period)
	return args.Error(0)
}

func (m *MockConnection) CreateSimulatedObject(containerTitle string, initPos SIMCONNECT_DATA_INITPOSITION, requestID uint32) error {
	args := m.Called(containerTitle, initPos, requestID)
	return args.Error(0)
}

func (m *MockConnection) GetNextDispatch() (*SIMCONNECT_RECV, bool) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*SIMCONNECT_RECV), args.Bool(1)
}
