package testutil

import (
	"atc_freq/internal/sim"
	"unsafe"
)

// FacilityDataBuffer is a struct that holds both the header and inline data
// This mimics how SimConnect returns data with a flexible array member
type FacilityDataBuffer struct {
	Pad_cgo_0             [12]byte // SIMCONNECT_RECV header
	UserRequestId         uint32
	UniqueRequestId       uint32
	ParentUniqueRequestId uint32
	Type                  uint32
	IsListItem            uint32
	ItemIndex             uint32
	ListSize              uint32
	Data                  sim.FACILITY_FREQUENCY_DATA // Inline frequency data
}

// CreateFacilityDataResponse creates a mock SIMCONNECT_RECV for facility data
func CreateFacilityDataResponse(freqType int32, frequency int32, name string) *sim.SIMCONNECT_RECV {
	var nameBytes [64]byte
	copy(nameBytes[:], name)

	buf := &FacilityDataBuffer{
		UserRequestId: sim.REQUEST_ID,
		Type:          sim.SIMCONNECT_FACILITY_DATA_FREQUENCY,
		Data: sim.FACILITY_FREQUENCY_DATA{
			TYPE:      freqType,
			FREQUENCY: frequency,
			NAME:      nameBytes,
		},
	}

	// Set DwID in the header (offset 8 bytes: DwSize(4) + DwVersion(4))
	*(*uint32)(unsafe.Pointer(&buf.Pad_cgo_0[8])) = sim.SIMCONNECT_RECV_ID_FACILITY_DATA

	return (*sim.SIMCONNECT_RECV)(unsafe.Pointer(buf))
}

// CreateFacilityDataEndResponse creates a mock SIMCONNECT_RECV for facility data end
func CreateFacilityDataEndResponse() *sim.SIMCONNECT_RECV {
	endData := &sim.SIMCONNECT_RECV_FACILITY_DATA_END{
		RequestId: sim.REQUEST_ID,
	}

	// The Pad_cgo_0 [12]byte contains the SIMCONNECT_RECV header
	// Set DwID in the header (offset 8 bytes: DwSize(4) + DwVersion(4))
	*(*uint32)(unsafe.Pointer(&endData.Pad_cgo_0[8])) = sim.SIMCONNECT_RECV_ID_FACILITY_DATA_END

	return (*sim.SIMCONNECT_RECV)(unsafe.Pointer(endData))
}