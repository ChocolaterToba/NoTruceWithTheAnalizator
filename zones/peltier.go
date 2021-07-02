package zones

import (
	"encoding/xml"
	"strconv"
	"strings"
)

type PeltierHoleInput struct {
	XMLName xml.Name `xml:"Hole"`
	Name    string   `xml:"name,attr"`
	X       string   `xml:"x,attr"`
	Y       string   `xml:"y,attr"`
	ZMin    string   `xml:"z_min,attr"`
	ZMax    string   `xml:"z_max,attr"`
	Volume  string   `xml:"volume,attr"`
}

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

func (holeInput *PeltierHoleInput) ToSection() (map[string]int16, error) {
	result := make(map[string]int16)
	var err error
	result["x"], err = parseNumericalVariables(holeInput.X)
	if err != nil {
		return nil, err
	}
	result["y"], err = parseNumericalVariables(holeInput.Y)
	if err != nil {
		return nil, err
	}
	result["z_min"], err = parseNumericalVariables(holeInput.ZMin)
	if err != nil {
		return nil, err
	}
	result["z_max"], err = parseNumericalVariables(holeInput.ZMax)
	if err != nil {
		return nil, err
	}
	result["volume"], err = parseNumericalVariables(holeInput.Volume)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type PeltierZoneInput struct {
	XMLName xml.Name           `xml:"Peltier"`
	Holes   []PeltierHoleInput `xml:"Hole"`
}

type Zone struct {
	Name     string
	Sections map[string]map[string]int16
}

func (zoneInput *PeltierZoneInput) ToZone() (*Zone, error) {
	result := new(Zone)
	result.Name = "Peltier"
	result.Sections = make(map[string]map[string]int16)
	for _, hole := range zoneInput.Holes {
		parsedHole, err := hole.ToSection()
		if err != nil {
			return nil, err
		}

		result.Sections[hole.Name] = parsedHole
	}

	return result, nil
}
