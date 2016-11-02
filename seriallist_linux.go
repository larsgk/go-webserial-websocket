package main

import (
	"bufio"
	serial "go.bug.st/serial"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func getPnPDetailsByUdevadm(path string) (*SerialPort, error) {
	out, err := exec.Command("udevadm", "info", "-q", "path", "-n", path).Output()
	if err != nil {
		log.Printf("Error creating exec.Command:  %v\n", err)
		return nil, err
	}

	syspath := strings.TrimSpace(string(out))

	cmd := exec.Command("udevadm", "info", "--query=property", "-p", syspath)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error creating StdoutPipe:  %v\n", err)
		return nil, err
	}

	var VID, PID uint16 = 0, 0
	var displayName string = ""

	scanner := bufio.NewScanner(cmdReader)
	stop := make(chan bool)
	go func() {
		for scanner.Scan() {
			// log.Printf("READ> %s\n", scanner.Text())
			keyValue := strings.Split(scanner.Text(), "=")
			if len(keyValue) > 1 {
				switch keyValue[0] {
				case "ID_MODEL", "ID_MODEL_FROM_DATABASE":
					if len(displayName) == 0 {
						displayName = keyValue[1]
					}
				case "ID_VENDOR_ID":
					val, _ := strconv.ParseUint(strings.Replace(keyValue[1], "0x", "", 1), 16, 16)
					VID = uint16(val)
				case "ID_MODEL_ID":
					val, _ := strconv.ParseUint(strings.Replace(keyValue[1], "0x", "", 1), 16, 16)
					PID = uint16(val)
				}

			}
		}
		stop <- true
	}()

	err = cmd.Start()
	if err != nil {
		log.Printf("Error starting cmd:  %v\n", err)
		return nil, err
	}

	<-stop
	err = cmd.Wait()
	if err != nil {
		log.Printf("Error waiting for cmd:  %v\n", err)
		return nil, err
	}

	return &SerialPort{Path: path, VendorId: VID, ProductId: PID, DisplayName: displayName}, err
}

func GetSerialPortList() ([]SerialPort, error) {
	serialPorts := []SerialPort{}

	paths, err := serial.GetPortsList()

	if err == nil {
		for _, path := range paths {
			serialPort, err := getPnPDetailsByUdevadm(path)
			if err == nil {
				serialPorts = append(serialPorts, *serialPort)
			}
		}
	}

	return serialPorts, err
}
