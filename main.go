package main

import (
	"bufio"
	"fmt"
	"os"
	"whatever/customError"
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

	fmt.Println("Ready to receive input")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		commandName := scanner.Text()
		if commandName == "" { // Empty command - end of input
			break
		}

		//TODO: prefixes
		neededCommand, found := commands[commandName]
		if !found {
			fmt.Printf("Error: %s\n", customError.CommandNotFoundError)
			continue
		}

		for _, subcommand := range neededCommand.Subcommands {
			pack := command.ToPackage(subcommand.Commands)
			errOrCode := boards.SendPackage(portName, pack)
			switch errOrCode.(type) {
			case error:
				fmt.Printf("Error: %s\n", errOrCode.(error))
			case int:
				fmt.Printf("Result code: %d\n", errOrCode.(int))
			}
		}
	}

	fmt.Println("Exiting program")
}
