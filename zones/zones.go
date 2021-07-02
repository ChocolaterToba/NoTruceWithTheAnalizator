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
	XMLName     xml.Name         `xml:"Zones"`
	PeltierZone PeltierZoneInput `xml:"Peltier"`
}

func (zonesInput *ZonesInput) ToZones() (map[string]*Zone, error) {
	result := make(map[string]*Zone)
	var err error
	result["Peltier"], err = zonesInput.PeltierZone.ToZone()
	if err != nil {
		return nil, err
	}

	return result, nil
}

type ZonesAndLogic struct {
	XMLName xml.Name `xml:"Config"`
	Zones   ZonesInput
	Logics  LogicsInput
}

func (zonesAndLogic *ZonesAndLogic) ToCommands() (map[string]*Command, error) {
	zones, err := zonesAndLogic.Zones.ToZones()
	if err != nil {
		return nil, err
	}

	return zonesAndLogic.Logics.ToCommands(zones)
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
