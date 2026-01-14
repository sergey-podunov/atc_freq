package sim

import (
	"atc_freq/internal/helpers"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	S_OK = 0
)

const (
	DEFINE_ID  = 0x1001
	REQUEST_ID = 0x2001
)

type DllConnection struct {
	dll *windows.DLL

	open                    *windows.Proc
	close                   *windows.Proc
	addToFacilityDefinition *windows.Proc
	requestFacilityData     *windows.Proc
	getNextDispatch         *windows.Proc

	handler uintptr
}

func NewConnection() (*DllConnection, error) {
	dll, err := windows.LoadDLL("SimConnect.dll")
	if err != nil {
		return nil, fmt.Errorf("load Connection.dll: %w", err)
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

	return &DllConnection{
		dll:                     dll,
		open:                    open,
		close:                   closeP,
		addToFacilityDefinition: addDef,
		requestFacilityData:     reqFac,
		getNextDispatch:         getDisp,
	}, nil
}

func (connection *DllConnection) Open(name string) error {
	namePtr, _ := helpers.CString(name)
	r1, _, _ := connection.open.Call(
		uintptr(unsafe.Pointer(&connection.handler)),
		uintptr(unsafe.Pointer(namePtr)),
		0, 0, 0, 0,
	)
	if int32(r1) != S_OK || connection.handler == 0 {
		return fmt.Errorf("SimConnect_Open failed HRESULT=0x%08X", uint32(r1))
	}
	return nil
}

func (connection *DllConnection) AddField(field string, defineID uint32) error {
	fptr, err := helpers.CString(field)
	if err != nil {
		return err
	}
	handlerResult, _, _ := connection.addToFacilityDefinition.Call(
		connection.handler,
		uintptr(defineID),
		uintptr(unsafe.Pointer(fptr)),
	)
	if int32(handlerResult) != S_OK {
		return fmt.Errorf("AddToFacilityDefinition(%q) failed HRESULT=0x%08X", field, uint32(handlerResult))
	}
	return nil
}

func (connection *DllConnection) RequestFacilityData(icao string, region string, defineID uint32, requestID uint32) error {
	icaoPtr, _ := helpers.CString(icao)
	regionPtr, _ := helpers.CString(region)

	handlerResult, _, _ := connection.requestFacilityData.Call(
		connection.handler,
		uintptr(defineID),
		uintptr(requestID),
		uintptr(unsafe.Pointer(icaoPtr)),
		uintptr(unsafe.Pointer(regionPtr)),
	)
	if int32(handlerResult) != S_OK {
		return fmt.Errorf("RequestFacilityData failed HRESULT=0x%08X", uint32(handlerResult))
	}

	return nil
}

func (connection *DllConnection) Close() {
	//todo connection.close.Call(connection.handler)
}

func (connection *DllConnection) GetNextDispatch() (*SIMCONNECT_RECV, bool) {
	var ppData *SIMCONNECT_RECV
	var cbData uint32

	handlerResult, _, _ := connection.getNextDispatch.Call(
		connection.handler,
		uintptr(unsafe.Pointer(&ppData)),
		uintptr(unsafe.Pointer(&cbData)),
	)

	if int32(handlerResult) != S_OK || ppData == nil {
		return nil, false
	}

	return ppData, true
}
