package sim

type ConnectionInterface interface {
	Open(name string) error
	Close()
	addField(field string, defineID uint32) error
	RequestFacilityData(icao string, region string, defineID uint32, requestID uint32) error
	GetNextDispatch() (*SIMCONNECT_RECV, bool)
}
