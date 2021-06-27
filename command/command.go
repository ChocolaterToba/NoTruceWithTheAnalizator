package command

import "whatever/crc"

type Command struct {
	id         byte
	command    byte
	dataFirst  byte
	dataSecond byte
}

func NewCommand(id byte, command byte, dataFirst byte, dataSecond byte) *Command {
	return &Command{
		id:         id,
		command:    command,
		dataFirst:  dataFirst,
		dataSecond: dataSecond,
	}
}

func (command *Command) AsBytes() []byte {
	result := make([]byte, 4)
	result[0] = command.id
	result[1] = command.command
	result[2] = command.dataFirst
	result[3] = command.dataSecond
	return result
}

func ToPackage(commands []*Command) []byte {
	result := make([]byte, len(commands)*4, 0)
	for _, command := range commands {
		result = append(result, command.AsBytes()...)
	}

	result = append(result, crc.Checksum(result)) // TODO: Check later
	return result
}
