package crc

import "github.com/go-daq/crc8"

var crcTable *crc8.Table = crc8.MakeTable(7)

func Checksum(bytes []byte) uint8 {
	return crc8.Checksum(bytes, crcTable)
}
