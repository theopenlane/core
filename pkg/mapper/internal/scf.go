package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	md "github.com/nao1215/markdown"

	"github.com/xuri/excelize/v2"
)

type ControlHeader string
type ControlValue string
type Framework string
type FrameworkControlID string
type SCFControlID string
type EvidenceRequestHeader string
type EvidenceRequestValue string

type Control map[ControlHeader]ControlValue
type SCFControls map[SCFControlID]Control
type EvidenceRequest map[EvidenceRequestHeader]EvidenceRequestValue
type EvidenceRequestID string

type ControlMapping map[Framework][]FrameworkControlID

func (c ControlMapping) MapsToControls() bool {
	for _, mappings := range c {
		if len(mappings) > 0 {
			return true
		}
	}
	return false
}

type SCFControlMappings map[SCFControlID]ControlMapping

const Description = "Description"
const ComplianceMethods = "Compliance Methods"
const ControlQuestions = "Control Questions"
const NotPerformed = "Not Performed"
const PerformedInternally = "Performed Informally"
const PlannedAndTracked = "Planned & Tracked"
const WellDefined = "Well Defined"
const QuantitativelyControlled = "Quantitatively Controlled"
const ContinuouslyImproving = "Continuously Improving"
const EvidenceRequestList = "Evidence Request List"

var SCFColumnMapping = map[string]ControlHeader{
	Description:              "Secure Controls Framework (SCF) Control Description",
	ControlQuestions:         "SCF Control Question",
	NotPerformed:             "C|P-CMM 0 Not Performed",
	PerformedInternally:      "C|P-CMM 1 Performed Informally",
	PlannedAndTracked:        "C|P-CMM 2 Planned & Tracked",
	WellDefined:              "C|P-CMM 3 Well Defined",
	QuantitativelyControlled: "C|P-CMM 4 Quantitatively Controlled",
	ContinuouslyImproving:    "C|P-CMM 5 Continuously Improving",
	EvidenceRequestList:      "Evidence Request List (ERL) #",
}

var SupportedFrameworks = map[Framework]ControlHeader{
	"SOC 2":     "AICPA TSC 2017 (with 2022 revised POF)",
	"GDPR":      "EMEA EU GDPR",
	"ISO 27001": "ISO 27001 v2022",
	"ISO 27002": "ISO 27002 v2022",
	// "ISO 27701":   "ISO 27701  v2019",
	"NIST 800-53": "NIST 800-53B rev5 (moderate)",
	// "HIPAA":       "US HIPAA",
}

var SCFControlFamilyMapping = map[string]string{
	"AAT": "Artificial and Autonomous Technology",
	"AST": "Asset Management",
	"BCD": "Business Continuity & Disaster Recovery",
	"CAP": "Capacity & Performance Planning",
	"CHG": "Change Management",
	"CLD": "Cloud Security",
	"CFG": "Configuration Management",
	"CPL": "Compliance",
	"CRY": "Cryptographic Protections",
	"DCH": "Data Classification & Handling",
	"EMB": "Embedded Technology",
	"END": "Endpoint Security",
	"GOV": "Cybersecurity & Data Privacy Governance",
	"HRS": "Human Resources Security",
	"IAO": "Information Assurance",
	"IAC": "Identification & Authentication",
	"IRO": "Incident Response",
	"MON": "Continuous Monitoring",
	"MNT": "Maintenance",
	"MDM": "Mobile Device Management",
	"NET": "Network Security",
	"OPS": "Security Operations",
	"PES": "Physical & Environmental Security",
	"PRI": "Data Privacy",
	"PRM": "Project & Resource Management",
	"RSK": "Risk Management",
	"SAT": "Security Awareness & Training",
	"SEA": "Secure Engineering & Architecture",
	"TDA": "Technology Development & Acquisition",
	"THR": "Threat Management",
	"TPM": "Third-Party Management",
	"VPM": "Vulnerability & Patch Management",
	"WEB": "Web Security",
}

func ReturnSCFControls(url string, getFile bool) (SCFControls, error) {
	controls := map[SCFControlID]Control{}
	if getFile {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
		out, err := os.Create("pkg/mapper/scf2.xlsx")
		if err != nil {
			return nil, err
		}

		defer out.Close()

		io.Copy(out, resp.Body)
	}

	f, err := excelize.OpenFile("pkg/mapper/scf2.xlsx")
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}()

	rows, err := f.GetRows("SCF 2024.3")
	if err != nil {
		return nil, err
	}

	headers := []ControlHeader{}
	for idx, row := range rows {
		if idx == 0 {
			for _, header := range row {
				headers = append(headers, ControlHeader(strings.ReplaceAll(header, "\n", " ")))
			}
		} else {
			scfControlID := fmt.Sprintf("%s - %s", row[2], strings.TrimSpace(strings.ReplaceAll(row[1], "\n", " ")))
			control := Control{}

			for idx, val := range row {
				control[headers[idx]] = ControlValue(strings.ReplaceAll(val, "▪", "-"))
			}
			controls[SCFControlID(scfControlID)] = control
		}
	}

	file, err := json.MarshalIndent(controls, "", " ")
	if err != nil {
		return controls, err
	}

	err = os.WriteFile("pkg/mapper/scf.json", file, 0644)
	if err != nil {
		return controls, err
	}

	return controls, nil
}

func ParseEvidenceRequest() (map[EvidenceRequestID]EvidenceRequest, error) {
	evidenceRequests := map[EvidenceRequestID]EvidenceRequest{}
	f, err := excelize.OpenFile("pkg/mapper/scf2.xlsx")
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}()

	rows, err := f.GetRows("Evidence Request List 2024.3")
	if err != nil {
		return nil, err
	}

	headers := []EvidenceRequestHeader{}

	for idx, row := range rows {
		if idx == 0 {
			for _, header := range row {
				headers = append(headers, EvidenceRequestHeader(header))
			}
		} else {
			evidenceRequestID := EvidenceRequestID(row[0])
			evidenceRequest := EvidenceRequest{}
			for idx, val := range row {
				evidenceRequest[headers[idx]] = EvidenceRequestValue(val)
			}
			evidenceRequests[evidenceRequestID] = evidenceRequest
		}
	}

	file, err := json.MarshalIndent(evidenceRequests, "", " ")
	if err != nil {
		return evidenceRequests, err
	}

	err = os.WriteFile("pkg/mapper/evidence.json", file, 0644)
	if err != nil {
		return evidenceRequests, err
	}
	return evidenceRequests, nil
}

func GenerateSCFMarkdown(scfControl Control, scfControlID SCFControlID, controlMapping ControlMapping) error {
	filename := fmt.Sprintf("scf/%s.md", safeFileName(string(scfControlID)))
	f, err := os.Create(filename)

	if err != nil {
		return err
	}

	description := string(scfControl[SCFColumnMapping[Description]])
	doc := md.NewMarkdown(f).
		H1(fmt.Sprintf("SCF - %s", string(scfControlID))).
		PlainText(description).
		H2("Mapped framework controls")

	orderedFrameworks := []string{}
	for framework := range controlMapping {
		orderedFrameworks = append(orderedFrameworks, string(framework))
	}

	slices.Sort(orderedFrameworks)

	for _, framework := range orderedFrameworks {
		frameworkControlIDs := controlMapping[Framework(framework)]
		fcids := []string{}

		for _, fcid := range frameworkControlIDs {
			link := fmt.Sprintf("[%s](../%s/%s.md)", string(fcid), safeFileName(string(framework)), safeFileName(string(fcid)))
			if framework == "GDPR" {
				articleParts := strings.Split(string(fcid), ".")
				if len(articleParts) == 2 {
					subArticle := strings.ReplaceAll(string(fcid), "Art", "Article")
					subArticle = strings.ReplaceAll(subArticle, ".", "")
					subArticle = strings.ReplaceAll(subArticle, " ", "-")
					link = fmt.Sprintf("[%s](../%s/%s.md#%s)", string(fcid), safeFileName(string(framework)), safeFileName(articleParts[0]), url.QueryEscape(subArticle))
				}
			} else if framework == "ISO 27001" || framework == "ISO 27002" {
				annex := FCIDToAnnex(Framework(framework), string(fcid))
				if strings.HasPrefix(annex, "A") {
					annexParts := strings.Split(annex, ".")
					annexLink := fmt.Sprintf("a-%s", annexParts[1])
					annexTarget := safeFileName(annex)
					link = fmt.Sprintf("[%s](../%s/%s.md#%s)", annex, safeFileName(framework), annexLink, annexTarget)
				} else {
					requirementParts := strings.Split(annex, ".")
					requirementLink := fmt.Sprintf("%s", requirementParts[0])
					requirementTarget := safeFileName(annex)
					link = fmt.Sprintf("[%s](../%s/%s.md#%s)", annex, safeFileName(framework), requirementLink, requirementTarget)
				}
			} else if framework == "NIST 800-53" {
				toLink := strings.ReplaceAll(strings.ReplaceAll(string(fcid), ")", ""), "(", "-")
				link = fmt.Sprintf("[%s](../nist80053/%s.md)", string(fcid), safeFileName(toLink))
			}

			found := false
			for _, fcid := range fcids {
				if fcid == link {
					found = true
				}
			}

			if !found {
				fcids = append(fcids, link)
			}
		}

		if len(fcids) > 0 {
			slices.Sort(fcids)
			doc.H3(string(framework)).
				BulletList(fcids...).
				LF()
		}
	}

	doc.H2("Evidence request list").
		PlainText(string(scfControl[SCFColumnMapping[EvidenceRequestList]])).
		LF()
	doc.H2("Control questions").
		PlainText(string(scfControl[SCFColumnMapping[ControlQuestions]])).
		LF()
	doc.H2("Compliance methods").
		PlainText(string(scfControl[SCFColumnMapping[ComplianceMethods]])).
		LF()
	doc.H2("Control maturity").
		H3("Not performed").
		PlainText(fixControlQuestions(string(scfControl[SCFColumnMapping[NotPerformed]]))).
		LF()
	doc.H3("Performed internally").
		PlainText(fixControlQuestions(string(scfControl[SCFColumnMapping[PerformedInternally]]))).
		LF()
	doc.H3("Planned and tracked").
		PlainText(fixControlQuestions(string(scfControl[SCFColumnMapping[PlannedAndTracked]]))).
		LF()
	doc.H3("Well defined").
		PlainText(fixControlQuestions(string(scfControl[SCFColumnMapping[WellDefined]]))).
		LF()
	doc.H3("Quantitatively controlled").
		PlainText(fixControlQuestions(string(scfControl[SCFColumnMapping[QuantitativelyControlled]]))).
		LF()
	doc.H3("Continuously improving").
		PlainText(fixControlQuestions(string(scfControl[SCFColumnMapping[ContinuouslyImproving]]))).
		LF()

	doc.Build()
	err = generateMetadata(filename, "SCF", string(scfControlID), "", description)
	if err != nil {
		return err
	}
	return nil
}

func fixControlQuestions(input string) string {
	return strings.ReplaceAll(strings.ReplaceAll(input, "•	", "- "), "<br>", "\n")
}

func GetComplianceControlMappings(controls SCFControls) SCFControlMappings {
	controlMappings := map[SCFControlID]ControlMapping{}

	for controlID, control := range controls {
		controlMapping := ControlMapping{}

		for framework, header := range SupportedFrameworks {
			fcids := strings.Split(string(control[header]), "\n")
			frameworkControlIDs := []FrameworkControlID{}

			for _, fcid := range fcids {
				frameworkControlIDs = append(frameworkControlIDs, FrameworkControlID(fcid))
			}

			controlMapping[framework] = frameworkControlIDs
			if len(controlMapping[framework]) == 1 && controlMapping[framework][0] == "" {
				controlMapping[framework] = []FrameworkControlID{}
			}
		}

		if controlMapping.MapsToControls() {
			controlMappings[controlID] = controlMapping
		}
	}

	return controlMappings
}

func GenerateSCFIndex(scfControlMappings SCFControlMappings, scfControls SCFControls) error {
	f, err := os.Create("pkg/mapper/scf/index.md")
	if err != nil {
		return err
	}

	doc := md.NewMarkdown(f).
		H1("SCF Controls")

	controlIDs := []string{}
	for scfControlID := range scfControlMappings {
		controlIDs = append(controlIDs, string(scfControlID))
	}

	slices.Sort(controlIDs)
	controlLinks := []string{}
	lastControlFamily := ""

	for _, controlID := range controlIDs {
		family := ""
		for fam := range SCFControlFamilyMapping {
			if strings.HasPrefix(controlID, fam) {
				family = fam
			}
		}

		if family != lastControlFamily {
			if lastControlFamily != "" {
				doc.BulletList(controlLinks...)
			}
			lastControlFamily = family
			doc.H2(fmt.Sprintf("%s - %s", family, SCFControlFamilyMapping[family]))
			controlLinks = []string{fmt.Sprintf("[%s](%s.md)", controlID, safeFileName(string(controlID)))}
		} else {
			controlLinks = append(controlLinks, fmt.Sprintf("[%s](%s.md)", controlID, safeFileName(string(controlID))))
		}
	}

	doc.Build()

	return nil
}

var BadCharacters = []string{
	"../",
	"<!--",
	"-->",
	"<",
	">",
	"'",
	"\"",
	"/",
	"&",
	"$",
	"#",
	"{", "}", "[", "]", "=",
	";", "?", "%20", "%22",
	"%3c", // <
	"%253",
}
