package main

import (
	"fmt"
	"whatever/zones"
)

func main() {
	// ports, _ := enumerator.GetDetailedPortsList()
	// // Parse xml
	// // get commands from it
	// // parse usb ports
	// // wait for console input
	// // when input is received, find corresponding set of commands
	// // turn it into bytes
	// // send them
	// // add response to some file, idk
	// fmt.Println(len(ports))
	// fmt.Println(ports[1].Name)
	// fmt.Println(crc.Checksum([]byte("1234")))
	commands, err := zones.ParseXML("input.xml")
	fmt.Println(err)
	fmt.Println(commands)
	fmt.Println(commands["HOME"])
	fmt.Println(commands["HOME"].Subcommands["PeltierHome"].Commands)
}
