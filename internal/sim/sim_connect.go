package sim

import (
	"atc_freq/internal/helpers"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	S_OK = 0

	SIMCONNECT_RECV_ID_EXCEPTION         = 3 // Add this
	SIMCONNECT_RECV_ID_FACILITY_DATA     = 28
	SIMCONNECT_RECV_ID_FACILITY_DATA_END = 29

	SIMCONNECT_FACILITY_DATA_FREQUENCY = 3
)

const (
	DEFINE_ID  = 0x1001
	REQUEST_ID = 0x2001
)

type SIMCONNECT_RECV struct {
	DwSize    uint32
	DwVersion uint32
	DwID      uint32
}

type SIMCONNECT_RECV_FACILITY_DATA struct {
	DwSize    uint32
	DwVersion uint32
	DwID      uint32

	UserRequestId         uint32
	UniqueRequestId       uint32
	ParentUniqueRequestId uint32
	Type                  uint32
	ItemIndex             uint32
	ListSize              uint32
	IsListItem            uint32 // Using uint32 for C++ BOOL alignment
}

type SIMCONNECT_RECV_FACILITY_DATA_END struct {
	DwSize    uint32
	DwVersion uint32
	DwID      uint32
	RequestId uint32
}

type simconnectDLL struct {
	dll *windows.DLL

	open                    *windows.Proc
	close                   *windows.Proc
	addToFacilityDefinition *windows.Proc
	requestFacilityData     *windows.Proc
	getNextDispatch         *windows.Proc
}

func loadSimConnect() (*simconnectDLL, error) {
	dll, err := windows.LoadDLL("SimConnect.dll")
	if err != nil {
		return nil, fmt.Errorf("load SimConnect.dll: %w", err)
	}

	mustProc := func(name string) (*windows.Proc, error) {
		p, e := dll.FindProc(name)
		if e != nil {
			dll.Release()
			return nil, fmt.Errorf("find proc %s: %w", name, e)
		}
		return p, nil
	}

	open, err := mustProc("SimConnect_Open")
	if err != nil {
		return nil, err
	}
	closeP, err := mustProc("SimConnect_Close")
	if err != nil {
		return nil, err
	}
	addDef, err := mustProc("SimConnect_AddToFacilityDefinition")
	if err != nil {
		return nil, err
	}
	reqFac, err := mustProc("SimConnect_RequestFacilityData")
	if err != nil {
		return nil, err
	}
	getDisp, err := mustProc("SimConnect_GetNextDispatch")
	if err != nil {
		return nil, err
	}

	return &simconnectDLL{
		dll:                     dll,
		open:                    open,
		close:                   closeP,
		addToFacilityDefinition: addDef,
		requestFacilityData:     reqFac,
		getNextDispatch:         getDisp,
	}, nil
}

func addField(field string, sc *simconnectDLL, hSimConnect uintptr) error {
	fptr, err := helpers.CString(field)
	if err != nil {
		return err
	}
	hr, _, _ := sc.addToFacilityDefinition.Call(
		hSimConnect,
		DEFINE_ID,
		uintptr(unsafe.Pointer(fptr)),
	)
	if int32(hr) != S_OK {
		return fmt.Errorf("AddToFacilityDefinition(%q) failed HRESULT=0x%08X", field, uint32(hr))
	}
	return nil
}
