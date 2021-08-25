package main

//#include <stdlib.h>
import "C"

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

// //export add
// func add(left int, right int) int {
// 	return left + right
// }

func runCommand(isLeftPort bool, commandName string) error {
	fmt.Printf("Trying to run command %s\n", commandName)
	commands, err := zones.ParseXML("input.xml")
	if err != nil {
		return fmt.Errorf("Could not parse xml, encountered error!")
	}

	leftPort, rightPort, err := findLeftRightPorts(commands)
	if err != nil {
		return err
	}
	if leftPort == nil && rightPort == nil {
		return fmt.Errorf("Error: boards not found")
	}

	var currPort *boards.Port
	switch isLeftPort {
	case true:
		if leftPort == nil {
			
			return fmt.Errorf("Error: left board not found")
		}
		currPort = leftPort
	case false:
		if rightPort == nil {
			return fmt.Errorf("Error: right board not found")
		}
		currPort = rightPort
	}

	switch commandName {
	case "LOAD_FROM_FILE":
		currPort.SendCommandsFromFile("example.txt")
	default:
		neededCommand, found := commands[commandName]
		if !found {
			return fmt.Errorf("Error: %s\n", customError.CommandNotFoundError)
		}

		currPort.SendCommand(neededCommand)
	}

	return nil
}

type LedType int

const (
	RedLed LedType = iota
	GreenLed
	BlueLed
)

func (ledType LedType) String() string {
	if ledType < RedLed || ledType > BlueLed {
		return "Unsupported LED type"
	}
    return [...]string{"LED_RED", "LED_GREEN", "LED_BLUE"}[ledType]
}

//export LedChange
func LedChange(isLeftPort bool, ledType LedType, turnOn bool) *C.char {
	ledCommand := ledType.String()
	if ledCommand == "Unsupported LED type" {
		return C.CString("Error: " + ledCommand)
	}

	switch turnOn {
	case true:
		ledCommand += "_ON"
	case false:
		ledCommand += "_OFF"
	}

	result := runCommand(isLeftPort, ledCommand).Error()
	resultAsCstring := C.CString(result)
	// defer C.free(unsafe.Pointer(resultAsCstring))
	return resultAsCstring
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
		default:
			fmt.Printf("Error: incorrect prefix, must be L or R: %s\n", commandLineArgs[0])
			continue
		}

		switch commandLineArgs[1] {
		case "LOAD_FROM_FILE":
			go currPort.SendCommandsFromFile("example.txt")
		case "INIT": // Reloads ports before executing
			leftPort, rightPort, err := findLeftRightPorts(commands)
			if err != nil {
				fmt.Println(err)
				return
			}
			if leftPort == nil && rightPort == nil {
				fmt.Println("Error: boards not found")
				return
			}

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
			default:
				fmt.Printf("Error: incorrect prefix, must be L or R: %s\n", commandLineArgs[0])
				continue
			}
			fallthrough // So that we actually execute INIT's subcommands
		default:
			neededCommand, found := commands[commandLineArgs[1]]
			if !found {
				fmt.Printf("Error: %s\n", customError.CommandNotFoundError)
				continue
			}

			go currPort.SendCommand(neededCommand)
		}
	}

	fmt.Println("Exiting program")
}
