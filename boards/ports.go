package boards

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"whatever/command"
	"whatever/crc"
	"whatever/zones"
)

type Port struct {
	name string
	busy bool
	mu   sync.Mutex
}

func NewPort(name string) *Port {
	return &Port{name: name}
}

func (port *Port) Lock() bool {
	port.mu.Lock()
	defer port.mu.Unlock()
	if port.busy {
		return false
	}

	port.busy = true
	return true
}

func (port *Port) Unlock() bool {
	port.mu.Lock()
	defer port.mu.Unlock()
	if !port.busy {
		return false
	}

	port.busy = false
	return true
}

func (port *Port) SendCommand(neededCommand *zones.Command) {
	if !port.Lock() {
		fmt.Println("Error: board is busy")
		return
	}

	defer port.Unlock()
	for _, subcommand := range neededCommand.Subcommands {
		pack := command.ToPackage(subcommand.Commands)
		errOrCode := port.SendPackage(pack)
		switch errOrCode.(type) {
		case error:
			fmt.Printf("Error: %s\n", errOrCode.(error))
		case int:
			fmt.Printf("Result code: %d\n", errOrCode.(int))
		}
	}
}

func (port *Port) SendPackage(pack []byte) interface{} {
	return sendPackage(port.name, pack)
}

func (port *Port) SendCommandsFromFile(filepath string) {
	commandFile, err := os.Open(filepath) // TODO: check for concurrency safety
	if err != nil {
		fmt.Println("Could not find file with commands")
		return
	}

	scanner := bufio.NewScanner(commandFile)
	packages := make([][]byte, 0)
	for scanner.Scan() {
		packageLine := scanner.Text()
		packageStrings := strings.Split(packageLine, " ")
		if len(packageStrings)%4 != 0 {
			fmt.Printf("Could not parse package: len not divisible by 4: %s\n", packageLine)
		}
		packageBytes := make([]byte, 0, len(packageStrings)+1)
		for i, packageString := range packageStrings {
			packageByte, err := strconv.ParseInt(packageString, 16, 16)
			if err != nil {
				fmt.Printf("Could not parse package's byte %d: \"%s\"\n", i, packageString)
				return
			}

			packageBytes = append(packageBytes, byte(packageByte))
		}

		packageBytes = append(packageBytes, crc.Checksum(packageBytes[:len(packageBytes)-1]))
		packages = append(packages, packageBytes)
	}

	fmt.Println("Packages parsed, starting execution")
	for _, pack := range packages {
		errOrCode := port.SendPackage(pack)
		switch errOrCode.(type) {
		case error:
			fmt.Printf("Error: %s\n", errOrCode.(error))
		case int:
			fmt.Printf("Result code: %d\n", errOrCode.(int))
		}
	}
}
