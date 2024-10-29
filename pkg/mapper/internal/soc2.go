package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	md "github.com/nao1215/markdown"
)

type ParsedDescription struct {
	Header string
	Body   string
}

type Requirement struct {
	ID          string `json:"Id"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Attributes  []struct {
		ItemID  string `json:"ItemId"`
		Section string `json:"Section"`
		Service string `json:"Service"`
		Type    string `json:"Type"`
	} `json:"Attributes"`
}

type FrameworkSummary struct {
	Framework    string        `json:"Framework"`
	Version      string        `json:"Version"`
	Provider     string        `json:"Provider"`
	Description  string        `json:"Description"`
	Requirements []Requirement `json:"Requirements"`
}

func GetSOC2Controls(url string, getFile bool) (FrameworkSummary, error) {
	frameworkSummary := FrameworkSummary{}

	if getFile {
		err := fetchFile(url, "soc2.json")
		if err != nil {
			return frameworkSummary, err
		}
	}

	soc2File, err := os.Open("soc2.json")
	if err != nil {
		return frameworkSummary, err
	}

	defer soc2File.Close()

	soc2Bytes, err := io.ReadAll(soc2File)

	if err != nil {
		return frameworkSummary, err
	}

	if err := json.Unmarshal(soc2Bytes, &frameworkSummary); err != nil {
		return frameworkSummary, err
	}

	return frameworkSummary, nil
}

// getFirstWord returns the first word of a string
func getFirstWord(input string) string {
	words := strings.Fields(input)
	if len(words) > 0 {
		return words[0]
	}

	return ""
}

func GenerateSOC2Markdown(requirement Requirement, scfControlMapping SCFControlMappings) error {
	socControlID := getFirstWord(requirement.Name)
	id := strings.ReplaceAll(requirement.ID, "cc_a", "a")
	id = strings.ReplaceAll(id, "cc_c", "c")

	filename := fmt.Sprintf("pkg/mapper/soc2/%s.md", safeFileName(id))
	f, err := os.Create(filename)

	if err != nil {
		return err
	}

	words := strings.Split(requirement.Name, " ")
	soc2id := words[0]

	doc := md.NewMarkdown(f).
		H1(fmt.Sprintf("SOC2 - %s", soc2id)).PlainText(md.Bold(strings.ReplaceAll(requirement.Name, soc2id+" ", "")))

	descriptions := parseSOC2Description(requirement.Description)

	for _, d := range descriptions {
		if d.Header == "" {
			doc.PlainText(d.Body)
		} else {
			doc.H2(d.Header).PlainText(d.Body)
		}
	}

	doc.H2("Mapped SCF controls")

	fcids := []string{}

	for scfID, controlMapping := range scfControlMapping {
		soc2FrameworkControlIDs := controlMapping["SOC 2"]
		for _, fcid := range soc2FrameworkControlIDs {
			if string(fcid) == socControlID {
				fcids = append(fcids, fmt.Sprintf("[%s](../scf/%s.md)", string(scfID), safeFileName(string(scfID))))
			}
		}
	}

	slices.Sort(fcids)

	doc.BulletList(fcids...)

	if err := doc.Build(); err != nil {
		return err
	}

	err = generateMetadata(filename, "SOC 2", id, "", requirement.Name)
	if err != nil {
		return err
	}

	return nil
}

func parseSOC2Description(description string) []ParsedDescription {
	descriptions := []ParsedDescription{}
	sentences := strings.Split(description, ". ")
	lastDescription := -1

	for _, sentence := range sentences {
		if strings.Contains(sentence, " - ") {
			parts := strings.Split(sentence, " - ")
			descriptions = append(descriptions, ParsedDescription{
				Header: parts[0],
				Body:   parts[1],
			})

			lastDescription++
		} else {
			if lastDescription == -1 {
				descriptions = append(descriptions, ParsedDescription{
					Header: "",
					Body:   sentence,
				})

				lastDescription++
			} else {
				descriptions[lastDescription].Body = fmt.Sprintf("%s. %s.", descriptions[lastDescription].Body, sentence)
			}
		}
	}

	return descriptions
}

func GenerateSOC2Index(soc2Framework FrameworkSummary) error {
	f, err := os.Create("pkg/mapper/soc2/index.md")
	if err != nil {
		return err
	}

	doc := md.NewMarkdown(f).
		H1("SOC2 Controls")
	controlLinks := []string{}

	for _, requirement := range soc2Framework.Requirements {
		id := strings.ReplaceAll(requirement.ID, "cc_a", "a")
		id = strings.ReplaceAll(id, "cc_c", "c")
		controlLinks = append(controlLinks, fmt.Sprintf("[%s](%s.md)", requirement.Name, safeFileName(id)))
	}

	slices.Sort(controlLinks)
	doc.BulletList(controlLinks...)

	if err := doc.Build(); err != nil {
		return err
	}

	return nil
}
