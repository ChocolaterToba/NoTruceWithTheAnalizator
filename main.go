package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"whatever/crc"
	"whatever/customError"
	"whatever/zones"

	"whatever/boards"

	"go.bug.st/serial/enumerator"
)

func findLeftRightPorts(commands map[string]*zones.Command) (*boards.Port, *boards.Port, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, nil, err
	}

	lrPack := []byte{1, 0, 0, 0, 0}
	lrPack[4] = byte(crc.Checksum(lrPack[:4]))

	var leftPort *boards.Port
	var rightPort *boards.Port
	for _, port := range ports {
		if port.VID == "0483" && port.PID == "5740" {
			lrPort := boards.NewPort(port.Name)
			errOrCode := lrPort.SendPackage(lrPack)
			switch errOrCode.(type) {
			case error:
				return nil, nil,
					fmt.Errorf("Error during l/r board distinguishing: %s\n", errOrCode.(error))
			case int:
				switch errOrCode.(int) {
				case 0:
					if leftPort != nil {
						return nil, nil,
							fmt.Errorf("Error during l/r board distinguishing: %s\n", "multiple left ports")
					}
					leftPort = lrPort
				case 1:
					if rightPort != nil {
						return nil, nil,
							fmt.Errorf("Error during l/r board distinguishing: %s\n", "multiple right ports")
					}
					rightPort = lrPort
				default:
					return nil, nil,
						fmt.Errorf("Error: unexpected l/r distinguishing response: %d\n", errOrCode.(int))
				}
			}
		}
	}

	return leftPort, rightPort, nil
}

func main() {
	commands, err := zones.ParseXML("input.xml")
	if err != nil {
		fmt.Println("Could not parse xml, encountered error!")
		fmt.Println(err)
		return
	}

	leftPort, rightPort, err := findLeftRightPorts(commands)
	if err != nil {
		fmt.Println(err)
		return
	}
	if leftPort == nil && rightPort == nil {
		fmt.Println("Error: boards not found")
		return
	}

	fmt.Println("Ready to receive input")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		commandLine := scanner.Text()
		if commandLine == "" { // Empty command - end of input
			break
		}
		commandLineArgs := strings.SplitN(commandLine, "_", 2)
		if len(commandLineArgs) != 2 {
			fmt.Println("Error: command is invalid - prefix missing")
			continue
		}

		var currPort *boards.Port
		switch commandLineArgs[0] {
		case "L":
			if leftPort == nil {
				fmt.Println("Error: left board not found")
				continue
			}
			currPort = leftPort
		case "R":
			if rightPort == nil {
				fmt.Println("Error: right board not found")
				continue
			}
			currPort = rightPort
		}

		neededCommand, found := commands[commandLineArgs[1]]
		if !found {
			switch commandLineArgs[1] {
			case "LOAD_FROM_FILE":
				go currPort.SendCommandsFromFile("example.txt")
			default:
				fmt.Printf("Error: %s\n", customError.CommandNotFoundError)
			}
			continue
		}

		go currPort.SendCommand(neededCommand)
	}

	fmt.Println("Exiting program")
}
