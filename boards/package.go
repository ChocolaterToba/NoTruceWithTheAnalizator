package boards

import (
	"fmt"
	"time"
	"whatever/crc"
	"whatever/customError"

	"github.com/tarm/serial"
)

const maxBusyLoops int = 5

const (
	FSM_GARBAGE int = -1
	FSM_OK_MIN  int = 0
	FSM_OK_MAX  int = 60000
	FSM_BUSY    int = 65534
	//Other errors may be added later if needed
)

const maxReadTimeout int = 100

func sendPackage(portSerial string, commands []byte) interface{} {
	config := &serial.Config{
		Name:        portSerial,
		Baud:        19200,
		ReadTimeout: time.Second * time.Duration(maxReadTimeout),
	}
	port, err := serial.OpenPort(config)
	if err != nil {
		return err
	}

	//Delete in prod, unnecessary
	fmt.Printf("Sending commands to port: % x\n", commands)

	err = sendInner(port, commands)
	if err != nil {
		return err
	}

	code, err := recvInner(port)
	if err == nil {
		return code
	}

	for busyCounter := 0; err == customError.BoardBusyError; busyCounter++ {
		if busyCounter > maxBusyLoops {
			break
		}

		_, err = recvInner(port)
		switch err {
		case customError.BoardReadyError:
			err = sendInner(port, commands)
			if err != nil {
				return err
			}

			code, err = recvInner(port)
			if err == nil {
				return code
			}
			return err
		case nil:
			return customError.BoardBusyError
		default:
			// Do nothing
		}
	}

	return err
}

func sendInner(port *serial.Port, commands []byte) error {
	totalSent := 0
	for totalSent < len(commands) {
		n, err := port.Write(commands[totalSent:])
		if err != nil {
			return err
		}

		totalSent += n
	}
	return nil
}

func recvInner(port *serial.Port) (int, error) {
	result := make([]byte, 0)

	startTime := time.Now()
	buff := make([]byte, 100)
	for time.Since(startTime) < time.Second*time.Duration(maxReadTimeout) {
		n, _ := port.Read(buff)
		if n == 0 {
			break
		}
		result = append(result, buff[:n]...)

		codeOrErr := parseBoardResponse(result) // TODO: only parse last something bytes
		switch codeOrErr.(type) {
		case int:
			//Delete printfs in prod, unnecessary
			fmt.Printf("Received response from port: % x\n", result)
			fmt.Printf("Received code from port: %d\n", codeOrErr.(int))
			return codeOrErr.(int), nil
		case error:
			if codeOrErr.(error) != customError.GarbageDataError {
				fmt.Printf("Received response from port: % x\n", result)
				return -1, codeOrErr.(error)
			}
		}
	}

	return -1, customError.ResponseTimeoutError
}

// returns error if code is not correct, int if it is correct
func parseBoardResponse(response []byte) interface{} {
	code := removeGarbage(response)
	switch {
	case code == FSM_GARBAGE:
		return customError.GarbageDataError
	case code == FSM_BUSY:
		return customError.BoardBusyError
	case code >= FSM_OK_MIN && code < FSM_OK_MAX:
		return code
	default:
		fmt.Printf("Unknown return code: %d\n", code)
		return fmt.Errorf("Unknown board error")
	}
}

func removeGarbage(response []byte) int {
	if len(response) < 5 {
		return -1
	}

	offset := 0
	for ; offset <= len(response)-5; offset++ {
		if crc.Checksum(response[offset:offset+4]) == response[offset+4] {
			break
		}
	}
	if offset > len(response)-5 { // Entire response was garbage
		return -1
	}

	return 256*int(response[offset+2]) + int(response[offset+3])
}
