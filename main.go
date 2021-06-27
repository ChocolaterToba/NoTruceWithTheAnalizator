package main

import (
	"fmt"

	"go.bug.st/serial/enumerator"
)

func main() {
	ports, _ := enumerator.GetDetailedPortsList()
	// Parse xml
	// get commands from it
	// parse usb ports
	// wait for console input
	// when input is received, find corresponding set of commands
	// turn it into bytes
	// send them
	// add response to some file, idk
	fmt.Println(len(ports))
	fmt.Println(ports[1].Name)
}
