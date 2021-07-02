package zones

import (
	"encoding/xml"
	"strconv"
	"strings"
)

func parseNumericalVariables(input string) (int16, error) {
	if strings.ContainsRune(input, '.') {
		parts := strings.SplitN(input, ".", 2)
		resultHigh, err := strconv.Atoi(parts[0])
		if err != nil {
			return -1, err
		}

		resultLow, err := strconv.Atoi(parts[1])
		if err != nil {
			return -1, err
		}

		return int16(resultHigh<<8 + resultLow), nil
	}

	result, err := strconv.ParseInt(input, 16, 16)
	if err != nil {
		return -1, err
	}
	return int16(result), nil
}

type SectionInput struct {
	XMLName xml.Name   `xml:"Section"`
	Name    string     `xml:"name,attr"`
	Vars    []VarInput `xml:"Var"`
}

type VarInput struct {
	XMLName xml.Name `xml:"Var"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}

func (sectionInput *SectionInput) ToSection() (map[string]int16, error) {
	result := make(map[string]int16)
	var err error
	for _, varInput := range sectionInput.Vars {
		result[varInput.Name], err = parseNumericalVariables(varInput.Value)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

type ZoneInput struct {
	XMLName  xml.Name       `xml:"Zone"`
	Name     string         `xml:"name,attr"`
	Sections []SectionInput `xml:"Section"`
}

type Zone struct {
	Name     string
	Sections map[string]map[string]int16
}

func (zoneInput *ZoneInput) ToZone() (*Zone, error) {
	result := new(Zone)
	result.Name = "Peltier"
	result.Sections = make(map[string]map[string]int16)
	for _, sectionInput := range zoneInput.Sections {
		section, err := sectionInput.ToSection()
		if err != nil {
			return nil, err
		}

		result.Sections[sectionInput.Name] = section
	}

	return result, nil
}
