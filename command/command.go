package command

import (
	"whatever/crc"
)

type Command struct {
	deviceID   byte
	commandID  byte
	dataFirst  byte
	dataSecond byte
}

func NewCommand(deviceID byte, commandID byte, dataFirst byte, dataSecond byte) *Command {
	return &Command{
		deviceID:   deviceID,
		commandID:  commandID,
		dataFirst:  dataFirst,
		dataSecond: dataSecond,
	}
}

func (command *Command) AsBytes() []byte {
	result := make([]byte, 4)
	result[0] = command.deviceID
	result[1] = command.commandID
	result[2] = command.dataFirst
	result[3] = command.dataSecond
	return result
}

func ToPackage(commands []*Command) []byte {
	result := make([]byte, 0, len(commands)*4)
	for _, command := range commands {
		result = append(result, command.AsBytes()...)
	}

	result = append(result, crc.Checksum(result)) // TODO: Check later
	return result
}
