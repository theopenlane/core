package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	md "github.com/nao1215/markdown"
)

type NIST80053ControlFamily struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Controls []NIST80053OSCALControl
}

type NIST80053Framework struct {
	Families []NIST80053ControlFamily
}

func GetNIST80053Controls(nisturl string, getFile bool) (NIST80053Framework, error) {
	fileName := "nist80053-v5.json"
	framework := NIST80053Framework{}

	if getFile {
		parsedURL, err := url.ParseRequestURI(nisturl)
		if err != nil {
			return framework, fmt.Errorf("invalid link: %w", err)
		}

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, parsedURL.String(), nil)
		if err != nil {
			return framework, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return framework, fmt.Errorf("failed to get ISO link: %w", err)
		}

		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				err = fmt.Errorf("failed to close response body: %w", cerr)
			}
		}()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return framework, err
		}

		oscal := NIST80053OSCAL{}

		err = json.Unmarshal(body, &oscal)
		if err != nil {
			return framework, err
		}

		for _, group := range oscal.Catalog.Groups {
			family := NIST80053ControlFamily{
				ID:    group.ID,
				Title: group.Title,
			}

			family.Controls = append(family.Controls, group.Controls...)

			framework.Families = append(framework.Families, family)
		}

		out, err := os.Create(fileName)
		if err != nil {
			return framework, err
		}

		defer out.Close()

		file, err := json.MarshalIndent(framework, "", " ")
		if err != nil {
			return framework, err
		}

		err = os.WriteFile(fileName, file, 0600) // nolint:mnd
		if err != nil {
			return framework, err
		}
	}

	gdprFile, err := os.Open(fileName)
	if err != nil {
		return framework, err
	}

	defer gdprFile.Close()

	nistBytes, err := io.ReadAll(gdprFile)
	if err != nil {
		return framework, err
	}

	if err := json.Unmarshal(nistBytes, &framework); err != nil {
		return framework, err
	}

	return framework, nil
}

func GenerateNIST80053Markdown(control NIST80053OSCALControl, scfControlMapping SCFControlMappings) error {
	filename := fmt.Sprintf("pkg/mapper/nist80053/%s.md", safeFileName(strings.ReplaceAll(control.ID, ".", "-")))

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	doc := md.NewMarkdown(f).
		H1(fmt.Sprintf("NIST 800-53v5 - %s - %s", strings.ToUpper(control.ID), control.Title))

	for _, part := range control.Parts {
		if part.Name == "statement" {
			for _, subPart := range part.Parts {
				prose := subPart.Prose
				regexPattern := `\{\{ insert: param, (\S*) \}\}`
				regex := regexp.MustCompile(regexPattern)
				matches := regex.FindStringSubmatch(prose)

				if len(matches) > 0 {
					for _, match := range matches {
						paramName := fmt.Sprintf("{{ insert: param, %s }}", match)
						paramValue := ""

						for _, param := range control.Params {
							if param.ID == match {
								paramValue = param.Label
							}
						}

						prose = strings.ReplaceAll(prose, paramName, fmt.Sprintf(`\[ Assignment: %s \]`, paramValue))
					}
				}

				doc.BulletList(prose)
			}
		} else if part.Name == "guidance" {
			doc.H2("Guidance").PlainText(strings.ReplaceAll(part.Prose, "\n", "\\n"))
		}
	}

	fcids := []string{}

	for scfID, controlMapping := range scfControlMapping {
		nistFrameworkControlIDs := controlMapping["NIST 800-53"]

		for _, fcid := range nistFrameworkControlIDs {
			controlIDToCompare := strings.ReplaceAll(strings.ToUpper(control.ID), ".", "-")
			toLink := strings.ReplaceAll(strings.ReplaceAll(string(fcid), ")", ""), "(", "-")

			if toLink == controlIDToCompare {
				fcids = append(fcids, fmt.Sprintf("[%s](../scf/%s.md)", string(scfID), safeFileName(string(scfID))))
			}
		}
	}

	slices.Sort(fcids)

	if len(fcids) > 0 {
		doc.H2("Mapped SCF controls")
		doc.BulletList(fcids...)
	}

	if err := doc.Build(); err != nil {
		return err
	}

	err = generateMetadata(filename, "NIST 800-53v5", control.ID, control.Title, "")
	if err != nil {
		return err
	}

	return nil
}

func GenerateNIST80053Index(nist80053Framework NIST80053Framework) error {
	f, err := os.Create("pkg/mapper/nist80053/index.md")
	if err != nil {
		return err
	}

	doc := md.NewMarkdown(f).
		H1("NIST 800-53 v5 Controls")

	for _, controls := range nist80053Framework.Families {
		doc.H2(fmt.Sprintf("%s - %s", strings.ToUpper(controls.ID), controls.Title))

		controlLinks := []string{}

		for _, control := range controls.Controls {
			controlLinks = append(controlLinks, fmt.Sprintf("[%s - %s](%s.md)", strings.ToUpper(control.ID), control.Title, safeFileName(string(control.ID))))

			for _, subControl := range control.Controls {
				subControlID := strings.ReplaceAll(subControl.ID, ".", "-")
				controlLinks = append(controlLinks, fmt.Sprintf("[%s - %s](%s.md)", strings.ToUpper(subControl.ID), subControl.Title, safeFileName(subControlID)))
			}
		}

		doc.BulletList(controlLinks...)
	}
	// slices.Sort(controlLinks)

	if err := doc.Build(); err != nil {
		return err
	}

	return nil
}

type NIST80053OSCALControl struct {
	ID     string `json:"id"`
	Class  string `json:"class"`
	Title  string `json:"title"`
	Params []struct {
		ID         string `json:"id"`
		Label      string `json:"label,omitempty"`
		Guidelines []struct {
			Prose string `json:"prose"`
		} `json:"guidelines,omitempty"`
		Select struct {
			HowMany string   `json:"how-many"`
			Choice  []string `json:"choice"`
		} `json:"select,omitempty"`
		Constraints []struct {
			Description string `json:"description"`
		} `json:"constraints,omitempty"`
	} `json:"params,omitempty"`
	Props []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		Class string `json:"class,omitempty"`
		Ns    string `json:"ns,omitempty"`
	} `json:"props"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Parts []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Parts []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Props []struct {
				Name    string `json:"name"`
				Ns      string `json:"ns,omitempty"`
				Value   string `json:"value"`
				Remarks string `json:"remarks,omitempty"`
			} `json:"props"`
			Prose string `json:"prose"`
			Parts []struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Props []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"props"`
				Prose string `json:"prose"`
				Parts []struct {
					ID    string `json:"id"`
					Name  string `json:"name"`
					Props []struct {
						Name  string `json:"name"`
						Value string `json:"value"`
					} `json:"props"`
					Prose string `json:"prose"`
				} `json:"parts,omitempty"`
			} `json:"parts,omitempty"`
		} `json:"parts,omitempty"`
		Prose string `json:"prose,omitempty"`
		Props []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
			Class string `json:"class"`
		} `json:"props,omitempty"`
		Links []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links,omitempty"`
	} `json:"parts"`
	Controls []NIST80053OSCALControl `json:"controls,omitempty"`
}

type NIST80053OSCAL struct {
	Catalog struct {
		UUID     string `json:"uuid"`
		Metadata struct {
			Title        string    `json:"title"`
			Published    time.Time `json:"published"`
			LastModified string    `json:"last-modified"`
			Version      string    `json:"version"`
			OscalVersion string    `json:"oscal-version"`
			Links        []struct {
				Href string `json:"href"`
				Rel  string `json:"rel"`
			} `json:"links"`
			Roles []struct {
				ID        string `json:"id"`
				Title     string `json:"title"`
				ShortName string `json:"short-name,omitempty"`
			} `json:"roles"`
			Parties []struct {
				UUID      string `json:"uuid"`
				Type      string `json:"type"`
				Name      string `json:"name"`
				ShortName string `json:"short-name"`
				Links     []struct {
					Href string `json:"href"`
					Rel  string `json:"rel"`
				} `json:"links"`
				EmailAddresses []string `json:"email-addresses,omitempty"`
				Addresses      []struct {
					Type       string   `json:"type"`
					AddrLines  []string `json:"addr-lines"`
					City       string   `json:"city"`
					State      string   `json:"state"`
					PostalCode string   `json:"postal-code"`
					Country    string   `json:"country"`
				} `json:"addresses,omitempty"`
			} `json:"parties"`
			ResponsibleParties []struct {
				RoleID     string   `json:"role-id"`
				PartyUuids []string `json:"party-uuids"`
			} `json:"responsible-parties"`
		} `json:"metadata"`
		Groups []struct {
			ID       string                  `json:"id"`
			Class    string                  `json:"class"`
			Title    string                  `json:"title"`
			Controls []NIST80053OSCALControl `json:"controls"`
		} `json:"groups"`
		BackMatter struct {
			Resources []struct {
				UUID     string `json:"uuid"`
				Title    string `json:"title,omitempty"`
				Citation struct {
					Text string `json:"text"`
				} `json:"citation,omitempty"`
				Rlinks []struct {
					Href string `json:"href"`
				} `json:"rlinks"`
				Description string `json:"description,omitempty"`
				Props       []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"props,omitempty"`
			} `json:"resources"`
		} `json:"back-matter"`
	} `json:"catalog"`
}
