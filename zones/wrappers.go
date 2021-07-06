package zones

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

type LogicsInput struct {
	XMLName  xml.Name       `xml:"Logic"`
	Commands []CommandInput `xml:"Command"`
}

func (logicsinput *LogicsInput) ToCommands(defaults map[string]*Zone) (map[string]*Command, error) {
	result := make(map[string]*Command) // TODO: fix referencing and stuff
	var err error
	for _, command := range logicsinput.Commands {
		result[command.Name], err = command.ToCommand(defaults)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

type ZonesInput struct {
	XMLName xml.Name    `xml:"Zones"`
	Zones   []ZoneInput `xml:"Zone"`
}

func (zonesInput *ZonesInput) ToZones() (map[string]*Zone, error) {
	result := make(map[string]*Zone)

	var err error
	for _, zoneInput := range zonesInput.Zones {
		result[zoneInput.Name], err = zoneInput.ToZone()
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

type InitSubcommandsInput struct {
	XMLName     xml.Name          `xml:"Init"`
	Subcommands []SubcommandInput `xml:"Subcommand"`
}

func (initSubcommandsInput *InitSubcommandsInput) ToCommand(defaults map[string]*Zone) (*Command, error) {
	initAsCommand := CommandInput{
		XMLName:     initSubcommandsInput.XMLName,
		Name:        "INIT",
		Subcommands: initSubcommandsInput.Subcommands,
	}

	return initAsCommand.ToCommand(defaults)
}

type ZonesAndLogic struct {
	XMLName         xml.Name `xml:"Config"`
	Zones           ZonesInput
	Logics          LogicsInput
	InitSubcommands InitSubcommandsInput
}

func (zonesAndLogic *ZonesAndLogic) ToCommands() (map[string]*Command, error) {
	zones, err := zonesAndLogic.Zones.ToZones()
	if err != nil {
		return nil, err
	}

	commands, err := zonesAndLogic.Logics.ToCommands(zones)
	if err != nil {
		return nil, err
	}

	initCommand, err := zonesAndLogic.InitSubcommands.ToCommand(zones)
	if err != nil {
		return nil, err
	}
	commands["Init"] = initCommand

	return commands, nil
}

func ParseXML(filepath string) (map[string]*Command, error) {
	xmlFile, err := os.Open(filepath) // TODO: check for concurrency safety
	if err != nil {
		return nil, err
	}
	xmlBytes, _ := ioutil.ReadAll(xmlFile)
	xmlFile.Close()

	zonesAndLogic := new(ZonesAndLogic)
	err = xml.Unmarshal(xmlBytes, zonesAndLogic)
	if err != nil {
		return nil, err
	}

	return zonesAndLogic.ToCommands()
}
