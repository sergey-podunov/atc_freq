package sim

type Connection interface {
	Open(name string) error
	Close()
	AddField(field string, defineID uint32) error
	RequestFacilityData(icao string, region string, defineID uint32, requestID uint32) error
	RequestWeatherObservation(icao string, requestID uint32) error
	GetNextDispatch() (*SIMCONNECT_RECV, bool)
}
