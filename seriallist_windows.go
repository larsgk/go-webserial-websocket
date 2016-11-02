package main

import (
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"log"
	"regexp"
	"strconv"
)

func init() {
	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		log.Fatal("Init error: ", err)
	}
}

func GetSerialPortList() ([]SerialPort, error) {
	// TODO: Handle potential panics better
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f", r)
		}
	}()

	objSWbemLocator, _ := oleutil.CreateObject("WbemScripting.SWbemLocator")
	defer objSWbemLocator.Release()

	objWMIService, _ := objSWbemLocator.QueryInterface(ole.IID_IDispatch)
	defer objWMIService.Release()

	serviceRaw, _ := oleutil.CallMethod(objWMIService, "ConnectServer")
	service := serviceRaw.ToIDispatch()
	defer service.Release()

	query := "SELECT * FROM Win32_PnPEntity WHERE ConfigManagerErrorCode = 0 and Name like '%(COM%'"
	queryResult, err := oleutil.CallMethod(service, "ExecQuery", query)

	serialPorts := []SerialPort{}

	if err != nil {
		log.Printf("Error from oleutil.CallMethod: ", err)
		return nil, err
	}

	result := queryResult.ToIDispatch()
	defer result.Release()

	countVar, _ := oleutil.GetProperty(result, "Count")
	count := int(countVar.Val)

	for i := 0; i < count; i++ {
		itemRaw, _ := oleutil.CallMethod(result, "ItemIndex", i)
		item := itemRaw.ToIDispatch()
		defer item.Release()

		displayName, _ := oleutil.GetProperty(item, "Name")

		re := regexp.MustCompile("\\((COM[0-9]+)\\)").FindAllStringSubmatch(displayName.ToString(), 1)

		var path string = ""

		if re != nil && len(re[0]) > 1 {
			path = re[0][1]
		}

		deviceId, _ := oleutil.GetProperty(item, "DeviceID")

		re = regexp.MustCompile("ID_(....)").FindAllStringSubmatch(deviceId.ToString(), 2)

		var VID, PID uint16 = 0, 0

		if re != nil && len(re) == 2 {
			if len(re[0]) > 1 {
				val, _ := strconv.ParseUint(re[0][1], 16, 16)
				VID = uint16(val)
			}
			if len(re[1]) > 1 {
				val, _ := strconv.ParseUint(re[1][1], 16, 16)
				PID = uint16(val)
			}
		}

		serialPorts = append(serialPorts, SerialPort{Path: path, VendorId: VID, ProductId: PID, DisplayName: displayName.ToString()})
	}

	return serialPorts, err
}
