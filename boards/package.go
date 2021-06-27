package boards

import (
	"fmt"
	"time"
	"whatever/crc"
	"whatever/errors"

	"github.com/tarm/serial"
)

const maxBusyLoops int = 5

const (
	FSM_GARBAGE int = -1
	FSM_OK      int = 0
	FSM_BUSY    int = 65534
	//Other errors may be added later if needed
)

const maxReadTimeout int = 100

func SendPackage(portSerial string, commands []byte) error {
	config := &serial.Config{
		Name:        portSerial,
		Baud:        19200,
		ReadTimeout: time.Second * time.Duration(maxReadTimeout),
	}
	port, err := serial.OpenPort(config)
	if err != nil {
		return err
	}

	err = sendInner(port, commands)
	if err != nil {
		return err
	}

	response, err := recvInner(port)
	if err != nil {
		return err
	}
	err = parseBoardResponse(response)

	for busyCounter := 0; err == errors.BoardBusyError; busyCounter++ {
		if busyCounter > maxBusyLoops {
			break
		}

		response, err = recvInner(port)
		if err != nil {
			return err
		}
		err = parseBoardResponse(response)

		if err == errors.BoardReadyError {
			err = sendInner(port, commands)
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

func recvInner(port *serial.Port) ([]byte, error) {
	result := make([]byte, 0)

	startTime := time.Now()
	buff := make([]byte, 100)
	for time.Since(startTime) < time.Second*time.Duration(maxReadTimeout) {
		n, err := port.Read(buff)
		if err != nil {
			return nil, err
		}

		if n == 0 {
			break
		}
		result = append(result, buff[:n]...)
	}

	return result, nil
}

func parseBoardResponse(response []byte) error {
	code := removeGarbage(response)
	switch code {
	case FSM_GARBAGE:
		return fmt.Errorf("Could not parse response due to insufficient data")
	case FSM_BUSY:
		return errors.BoardBusyError
	case FSM_OK:
		return nil
	default:
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
