package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"whatever/customError"
	"whatever/zones"

	"whatever/boards"
	"whatever/command"

	"go.bug.st/serial/enumerator"
)

type port struct {
	name string
	busy bool
}

func (currentPort *port) lock() bool {
	if currentPort.busy {
		return false
	}

	currentPort.busy = true
	return true
}

func (currentPort *port) unlock() bool {
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

	lrCommand, found := commands["REPLACE THIS"]
	if !found {
		return "", "", fmt.Errorf("Error: %s\n",
			"Could not find command for distinguishing between left/right boards")
	}
	lrSubcommand := lrCommand.Subcommands[0]
	lrPack := command.ToPackage(lrSubcommand.Commands)

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
