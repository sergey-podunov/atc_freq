package sim

import (
	"atc_freq/internal/helpers"
	"fmt"
	"strings"
	"time"
	"unsafe"
)

type FACILITY_FREQUENCY_DATA struct {
	TYPE      int32
	FREQUENCY int32    // Hz
	NAME      [64]byte // C char[64]
}

type AirportFrequency struct {
	Type     string
	TypeCode int32
	Name     string
	Hz       int
	MHz      float64
}

var freqTypeMap = map[int32]string{
	0:  "NONE",
	1:  "ATIS",
	2:  "MULTICOM",
	3:  "UNICOM",
	4:  "CTAF",
	5:  "GROUND",
	6:  "TOWER",
	7:  "CLEARANCE",
	8:  "APPROACH",
	9:  "DEPARTURE",
	10: "CENTER",
	11: "FSS",
	12: "AWOS",
	13: "ASOS",
	14: "CPT",
	15: "GCO",
}

type SimClient struct {
	sc *simconnectDLL
}

func NewSimClient() (*SimClient, error) {
	sc, err := loadSimConnect()
	if err != nil {
		return nil, err
	}
	return &SimClient{sc: sc}, nil
}

func (c *SimClient) GetAirportFrequencies(icao string, timeout time.Duration) ([]AirportFrequency, error) {
	icao = strings.ToUpper(strings.TrimSpace(icao))
	if icao == "" {
		return nil, fmt.Errorf("icao is empty")
	}

	// HANDLE hSimConnect;
	var hSimConnect uintptr

	// HRESULT SimConnect_Open(HANDLE* phSimConnect, LPCSTR name, HWND, DWORD, HANDLE, DWORD)
	namePtr, _ := helpers.CString("go-freq-client")
	r1, _, _ := c.sc.open.Call(
		uintptr(unsafe.Pointer(&hSimConnect)),
		uintptr(unsafe.Pointer(namePtr)),
		0, 0, 0, 0,
	)

	fmt.Printf("Open HRESULT: 0x%08X, Handle: %v\n", uint32(r1), hSimConnect)

	if int32(r1) != S_OK || hSimConnect == 0 {
		return nil, fmt.Errorf("SimConnect_Open failed HRESULT=0x%08X", uint32(r1))
	}
	defer c.sc.close.Call(hSimConnect)

	// Facility definition: OPEN AIRPORT -> OPEN FREQUENCY -> TYPE/FREQUENCY/NAME -> CLOSE -> CLOSE
	if err := addField("OPEN AIRPORT", c.sc, hSimConnect); err != nil {
		return nil, err
	}
	if err := addField("OPEN FREQUENCY", c.sc, hSimConnect); err != nil {
		return nil, err
	}
	if err := addField("TYPE", c.sc, hSimConnect); err != nil {
		return nil, err
	}
	if err := addField("FREQUENCY", c.sc, hSimConnect); err != nil {
		return nil, err
	}
	if err := addField("NAME", c.sc, hSimConnect); err != nil {
		return nil, err
	}
	if err := addField("CLOSE FREQUENCY", c.sc, hSimConnect); err != nil {
		return nil, err
	}
	if err := addField("CLOSE AIRPORT", c.sc, hSimConnect); err != nil {
		return nil, err
	}

	icaoPtr, _ := helpers.CString(icao)
	regionPtr, _ := helpers.CString("")

	// HRESULT SimConnect_RequestFacilityData(HANDLE, defineId, requestId, icao, region)
	hr, _, _ := c.sc.requestFacilityData.Call(
		hSimConnect,
		DEFINE_ID,
		REQUEST_ID,
		uintptr(unsafe.Pointer(icaoPtr)),
		uintptr(unsafe.Pointer(regionPtr)),
	)
	if int32(hr) != S_OK {
		return nil, fmt.Errorf("RequestFacilityData failed HRESULT=0x%08X", uint32(hr))
	}

	var out []AirportFrequency
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// HRESULT SimConnect_GetNextDispatch(HANDLE, SIMCONNECT_RECV** ppData, DWORD* pcbData)
		var ppData *SIMCONNECT_RECV
		var cbData uint32

		hr, _, _ := c.sc.getNextDispatch.Call(
			hSimConnect,
			uintptr(unsafe.Pointer(&ppData)),
			uintptr(unsafe.Pointer(&cbData)),
		)

		if int32(hr) != S_OK || ppData == nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		fmt.Printf("dispatch: %#v\n", ppData)

		switch ppData.DwID {
		case SIMCONNECT_RECV_ID_EXCEPTION:
			// Cast to SIMCONNECT_RECV_EXCEPTION and print dwException
			fmt.Printf("SimConnect Exception received! ID: %d\n", ppData.DwID)
		case SIMCONNECT_RECV_ID_FACILITY_DATA:
			// We will read fields by offset to avoid alignment issues
			// ppData is the start of the message
			basePtr := uintptr(unsafe.Pointer(ppData))

			// According to MSFS SDK:
			// Offset 0: Header (12 bytes: dwSize, dwVersion, dwID)
			// Offset 12: UserRequestId (uint32)
			// Offset 16: UniqueRequestId (uint32)
			// Offset 20: ParentUniqueRequestId (uint32)
			// Offset 24: Type (uint32)
			// Offset 28: ItemIndex (uint32)
			// Offset 32: ListSize (uint32)
			// Offset 36: IsListItem (uint32/BOOL)
			// Offset 40: DATA STARTS HERE

			userReqId := *(*uint32)(unsafe.Pointer(basePtr + 12))
			msgType := *(*uint32)(unsafe.Pointer(basePtr + 24))

			if userReqId != REQUEST_ID {
				continue
			}

			if msgType != SIMCONNECT_FACILITY_DATA_FREQUENCY {
				continue
			}

			fmt.Println("Found Frequency Data!")

			// Frequency data starts at offset 40
			dataPtr := unsafe.Pointer(basePtr + 40)
			freq := (*FACILITY_FREQUENCY_DATA)(dataPtr)

			name := helpers.TrimCString(freq.NAME[:])
			hz := int(freq.FREQUENCY)
			tcode := freq.TYPE
			tname := freqTypeMap[tcode]
			if tname == "" {
				tname = fmt.Sprintf("UNKNOWN_%d", tcode)
			}

			out = append(out, AirportFrequency{
				Type:     tname,
				TypeCode: tcode,
				Name:     name,
				Hz:       hz,
				MHz:      helpers.HzToMHz(hz),
			})

		case SIMCONNECT_RECV_ID_FACILITY_DATA_END:
			fmt.Println("got facility data end")
			end := (*SIMCONNECT_RECV_FACILITY_DATA_END)(unsafe.Pointer(ppData))
			if end.RequestId == REQUEST_ID {
				return out, nil
			}
		}
	}

	return out, fmt.Errorf("timeout waiting for facility data end (got %d frequencies so far)", len(out))
}
