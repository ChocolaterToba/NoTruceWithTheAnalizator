package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"whatever/crc"
	"whatever/customError"
	"whatever/zones"

	"whatever/boards"
	"whatever/command"

	"go.bug.st/serial/enumerator"
)

type port struct {
	name string
	busy bool
	mu   sync.Mutex
}

func (currentPort *port) lock() bool {
	currentPort.mu.Lock()
	defer currentPort.mu.Unlock()
	if currentPort.busy {
		return false
	}

	currentPort.busy = true
	return true
}

func (currentPort *port) unlock() bool {
	currentPort.mu.Lock()
	defer currentPort.mu.Unlock()
	if !currentPort.busy {
		return false
	}

	currentPort.busy = false
	return true
}

func (currentPort *port) sendCommand(neededCommand *zones.Command) {
	if !currentPort.lock() {
		fmt.Println("Error: board is busy")
		return
	}

	defer currentPort.unlock()
	for _, subcommand := range neededCommand.Subcommands {
		pack := command.ToPackage(subcommand.Commands)
		errOrCode := boards.SendPackage(currentPort.name, pack)
		switch errOrCode.(type) {
		case error:
			fmt.Printf("Error: %s\n", errOrCode.(error))
		case int:
			fmt.Printf("Result code: %d\n", errOrCode.(int))
		}
	}
}

func findLeftRightPorts(commands map[string]*zones.Command) (string, string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return "", "", err
	}

	lrPack := []byte{1, 0, 0, 0, 0}
	lrPack[4] = byte(crc.Checksum(lrPack[:4]))
	fmt.Println(lrPack)

	leftPort := ""
	rightPort := ""
	for _, port := range ports {
		if port.VID == "0483" && port.PID == "5740" {
			errOrCode := boards.SendPackage(port.Name, lrPack)
			switch errOrCode.(type) {
			case error:
				return "", "",
					fmt.Errorf("Error during l/r board distinguishing: %s\n", errOrCode.(error))
			case int:
				switch errOrCode.(int) {
				case 0:
					leftPort = port.Name
				case 1:
					rightPort = port.Name
				default:
					return "", "",
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

	leftPort := port{}
	rightPort := port{}
	leftPort.name, rightPort.name, err = findLeftRightPorts(commands)
	if err != nil {
		fmt.Println(err)
		return
	}

	if leftPort.name == "" {
		fmt.Println("Could not find left board!")
		return
	}
	if rightPort.name == "" {
		fmt.Println("Could not find right board!")
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

		//TODO: prefixes
		neededCommand, found := commands[commandLineArgs[1]]
		if !found {
			fmt.Printf("Error: %s\n", customError.CommandNotFoundError)
			continue
		}

		switch commandLineArgs[0] {
		case "L":
			go leftPort.sendCommand(neededCommand)
		case "R":
			go rightPort.sendCommand(neededCommand)
		}
	}

	fmt.Println("Exiting program")
}
