package sim

type Connection interface {
	Open(name string) error
	Close()
	AddField(field string, defineID uint32) error
	AddToDataDefinition(defineID uint32, datumName string, unitsName string, datumType uint32) error
	RequestFacilityData(icao string, region string, defineID uint32, requestID uint32) error
	RequestWeatherObservation(icao string, requestID uint32) error
	RequestCloudState(requestID uint32, minLat, minLon, minAlt, maxLat, maxLon, maxAlt float32) error
	RequestDataOnSimObjectType(requestID, defineID uint32, radius uint32, objectType uint32) error
	CreateSimulatedObject(containerTitle string, initPos SIMCONNECT_DATA_INITPOSITION, requestID uint32) error
	GetNextDispatch() (*SIMCONNECT_RECV, bool)
}
