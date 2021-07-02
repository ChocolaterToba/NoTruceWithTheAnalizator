package main

import (
	"fmt"
	"whatever/zones"

	"whatever/boards"
	"whatever/command"

	"go.bug.st/serial/enumerator"
)

func main() {
	commands, err := zones.ParseXML("input.xml")
	if err != nil {
		fmt.Println("Could not parse xml, encountered error!")
		fmt.Println(err)
		return
	}

	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		fmt.Println(err)
		return
	}

	portName := ""
	for _, port := range ports {
		if port.VID == "0483" && port.PID == "5740" {
			portName = port.Name
			break
		}
	}
	if portName == "" {
		fmt.Println("Could not find needed port!")
		//return
	}

	// wait for console input
	// when input is received, find corresponding set of commands
	// turn it into bytes
	// send them
	// add response to some file, idk

	// let's say, hypothetically, that user asked to launch command HOME
	neededCommand := commands["HOME"]
	for _, subcommand := range neededCommand.Subcommands {
		pack := command.ToPackage(subcommand.Commands)
		boards.SendPackage(portName, pack)
	}
}
