package zones

import (
	"encoding/xml"
	"strconv"
	"whatever/command"
	"whatever/customError"
)

type CommandInput struct {
	XMLName     xml.Name          `xml:"Command"`
	Name        string            `xml:"name,attr"`
	Subcommands []SubcommandInput `xml:"Subcommand"`
}

type Command struct {
	Name        string
	Subcommands map[string]*Subcommand
}

func (commandInput *CommandInput) ToCommand(defaults map[string]*Zone) (*Command, error) {
	result := new(Command)
	result.Name = commandInput.Name
	result.Subcommands = make(map[string]*Subcommand)
	for _, subcommandInput := range commandInput.Subcommands {
		newSubcommand, err := subcommandInput.ToSubcommand(defaults)
		if err != nil {
			return nil, err
		}
		result.Subcommands[newSubcommand.Name] = newSubcommand
	}
	return result, nil
}

type SubcommandInput struct {
	XMLName xml.Name `xml:"Subcommand"`
	Name    string   `xml:"name,attr"`
	//Default
	Zone  string      `xml:"zone,attr"`
	Codes []CodeInput `xml:"Code"`
}

type Subcommand struct {
	Name     string
	Commands []command.Command
}

func (subcommandInput *SubcommandInput) ToSubcommand(defaults map[string]*Zone) (*Subcommand, error) {
	result := new(Subcommand)
	result.Name = subcommandInput.Name
	zone, found := defaults[subcommandInput.Zone]
	if !found {
		return nil, customError.ZoneNotFoundError
	}

	result.Commands = make([]command.Command, 0, len(subcommandInput.Codes))
	for _, codeInput := range subcommandInput.Codes {
		command, err := codeInput.ToCommand(zone.Sections)
		if err != nil {
			return nil, err
		}

		result.Commands = append(result.Commands, *command)
	}

	return result, nil
}

type CodeInput struct {
	XMLName   xml.Name `xml:"Code"`
	Var       string   `xml:"var,attr"`
	Section   string   `xml:"section,attr"`
	DeviceID  string   `xml:"deviceID,attr"`
	CommandID string   `xml:",chardata"`
}

func (codeInput *CodeInput) ToCommand(sections map[string]map[string]int16) (*command.Command, error) {
	section, found := sections[codeInput.Section]
	if !found {
		return nil, customError.SectionNotFoundError
	}

	deviceID, err := strconv.Atoi(codeInput.DeviceID)
	if err != nil {
		return nil, err
	}

	commandID, err := strconv.Atoi(codeInput.CommandID)
	if err != nil {
		return nil, err
	}

	param, found := section[codeInput.Var]
	if !found {
		return nil, customError.ParamNotFoundError
	}

	return command.NewCommand(
			byte(int8(deviceID)), byte(int8(commandID)),
			byte(int8(param>>8)), byte(int8(param))),
		nil
}
