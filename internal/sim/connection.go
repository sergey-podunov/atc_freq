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

	open                       *windows.Proc
	close                      *windows.Proc
	addToFacilityDefinition    *windows.Proc
	addToDataDefinition        *windows.Proc
	requestFacilityData        *windows.Proc
	requestWeatherObservation  *windows.Proc
	requestCloudState          *windows.Proc
	requestDataOnSimObjectType *windows.Proc
	requestDataOnSimObject     *windows.Proc
	createSimulatedObject      *windows.Proc
	getNextDispatch            *windows.Proc

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
	addDataDef, err := mustProc("SimConnect_AddToDataDefinition")
	if err != nil {
		return nil, err
	}
	reqFac, err := mustProc("SimConnect_RequestFacilityData")
	if err != nil {
		return nil, err
	}
	reqWeather, err := mustProc("SimConnect_WeatherRequestObservationAtStation")
	if err != nil {
		return nil, err
	}
	reqCloudState, err := mustProc("SimConnect_WeatherRequestCloudState")
	if err != nil {
		return nil, err
	}
	reqDataOnSimObjectType, err := mustProc("SimConnect_RequestDataOnSimObjectType")
	if err != nil {
		return nil, err
	}
	reqDataOnSimObject, err := mustProc("SimConnect_RequestDataOnSimObject")
	if err != nil {
		return nil, err
	}
	createSimObject, err := mustProc("SimConnect_AICreateSimulatedObject")
	if err != nil {
		return nil, err
	}
	getDisp, err := mustProc("SimConnect_GetNextDispatch")
	if err != nil {
		return nil, err
	}

	fmt.Println("Connection initialized")
	return &DllConnection{
		dll:                        dll,
		open:                       open,
		close:                      closeP,
		addToFacilityDefinition:    addDef,
		addToDataDefinition:        addDataDef,
		requestFacilityData:        reqFac,
		requestWeatherObservation:  reqWeather,
		requestCloudState:          reqCloudState,
		requestDataOnSimObjectType: reqDataOnSimObjectType,
		requestDataOnSimObject:     reqDataOnSimObject,
		createSimulatedObject:      createSimObject,
		getNextDispatch:            getDisp,
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

func (connection *DllConnection) RequestWeatherObservation(icao string, requestID uint32) error {
	icaoPtr, _ := helpers.CString(icao)

	handlerResult, _, _ := connection.requestWeatherObservation.Call(
		connection.handler,
		uintptr(requestID),
		uintptr(unsafe.Pointer(icaoPtr)),
	)
	if int32(handlerResult) != S_OK {
		return fmt.Errorf("RequestWeatherObservation failed HRESULT=0x%08X", uint32(handlerResult))
	}

	return nil
}

func (connection *DllConnection) RequestCloudState(requestID uint32, minLat, minLon, minAlt, maxLat, maxLon, maxAlt float32) error {
	handlerResult, _, _ := connection.requestCloudState.Call(
		connection.handler,
		uintptr(requestID),
		uintptr(*(*uint32)(unsafe.Pointer(&minLat))),
		uintptr(*(*uint32)(unsafe.Pointer(&minLon))),
		uintptr(*(*uint32)(unsafe.Pointer(&minAlt))),
		uintptr(*(*uint32)(unsafe.Pointer(&maxLat))),
		uintptr(*(*uint32)(unsafe.Pointer(&maxLon))),
		uintptr(*(*uint32)(unsafe.Pointer(&maxAlt))),
		0, // dwFlags
	)
	if int32(handlerResult) != S_OK {
		return fmt.Errorf("RequestCloudState failed HRESULT=0x%08X", uint32(handlerResult))
	}

	return nil
}

func (connection *DllConnection) CreateSimulatedObject(containerTitle string, initPos SIMCONNECT_DATA_INITPOSITION, requestID uint32) error {
	titlePtr, err := helpers.CString(containerTitle)
	if err != nil {
		return err
	}

	handlerResult, _, _ := connection.createSimulatedObject.Call(
		connection.handler,
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(unsafe.Pointer(&initPos)),
		uintptr(requestID),
	)
	if int32(handlerResult) != S_OK {
		return fmt.Errorf("AICreateSimulatedObject failed HRESULT=0x%08X", uint32(handlerResult))
	}

	return nil
}

func (connection *DllConnection) Close() {
	//todo connection.close.Call(connection.handler)
}

func (connection *DllConnection) AddToDataDefinition(defineID uint32, datumName string, unitsName string, datumType uint32) error {
	datumPtr, err := helpers.CString(datumName)
	if err != nil {
		return err
	}
	unitsPtr, err := helpers.CString(unitsName)
	if err != nil {
		return err
	}

	handlerResult, _, _ := connection.addToDataDefinition.Call(
		connection.handler,
		uintptr(defineID),
		uintptr(unsafe.Pointer(datumPtr)),
		uintptr(unsafe.Pointer(unitsPtr)),
		uintptr(datumType),
		0, // fEpsilon (default)
		0, // DatumID (unused)
	)
	if int32(handlerResult) != S_OK {
		return fmt.Errorf("AddToDataDefinition(%q) failed HRESULT=0x%08X", datumName, uint32(handlerResult))
	}
	return nil
}

func (connection *DllConnection) RequestDataOnSimObjectType(requestID, defineID uint32, radius uint32, objectType uint32) error {
	handlerResult, _, _ := connection.requestDataOnSimObjectType.Call(
		connection.handler,
		uintptr(requestID),
		uintptr(defineID),
		uintptr(radius),
		uintptr(objectType),
	)
	if int32(handlerResult) != S_OK {
		return fmt.Errorf("RequestDataOnSimObjectType failed HRESULT=0x%08X", uint32(handlerResult))
	}
	return nil
}

func (connection *DllConnection) RequestDataOnSimObject(requestID, defineID, objectID, period uint32) error {
	handlerResult, _, _ := connection.requestDataOnSimObject.Call(
		connection.handler,
		uintptr(requestID),
		uintptr(defineID),
		uintptr(objectID),
		uintptr(period),
		0, // flags
		0, // origin
		0, // interval
		0, // limit
	)
	if int32(handlerResult) != S_OK {
		return fmt.Errorf("RequestDataOnSimObject failed HRESULT=0x%08X", uint32(handlerResult))
	}
	return nil
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
