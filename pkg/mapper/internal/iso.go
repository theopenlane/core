package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"

	md "github.com/nao1215/markdown"
)

type ISOControl struct {
	Ref     string `json:"ref"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

type ISODomain struct {
	Title    string       `json:"title"`
	Controls []ISOControl `json:"controls"`
}

type ISOFramework struct {
	Domains []ISODomain `json:"domains"`
}

func GetISOControls(standard Framework, isoLink string, getFile bool) (ISOFramework, error) {
	isoFramework := ISOFramework{}
	filename := fmt.Sprintf("pkg/mapper/%s.json", safeFileName(string(standard)))

	if getFile {
		parsedURL, err := url.ParseRequestURI(isoLink)
		if err != nil {
			return isoFramework, fmt.Errorf("invalid ISO link: %w", err)
		}

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, parsedURL.String(), nil)
		if err != nil {
			return isoFramework, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return isoFramework, fmt.Errorf("failed to get ISO link: %w", err)
		}

		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				err = fmt.Errorf("failed to close response body: %w", cerr)
			}
		}()

		out, err := os.Create(filename)
		if err != nil {
			return isoFramework, fmt.Errorf("failed to create file: %w", err)
		}

		defer func() {
			if cerr := out.Close(); cerr != nil {
				err = fmt.Errorf("failed to close file: %w", cerr)
			}
		}()

		if _, err := io.Copy(out, resp.Body); err != nil {
			return isoFramework, fmt.Errorf("failed to copy response body to file: %w", err)
		}
	}

	isoFile, err := os.Open(filename)
	if err != nil {
		return isoFramework, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		if cerr := isoFile.Close(); cerr != nil {
			err = fmt.Errorf("failed to close file: %w", cerr)
		}
	}()

	isoBytes, err := io.ReadAll(isoFile)
	if err != nil {
		return isoFramework, fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(isoBytes, &isoFramework); err != nil {
		return isoFramework, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return isoFramework, nil
}

func GenerateISOMarkdown(standard Framework, isoDomain ISODomain, scfControlMapping SCFControlMappings) error {
	if standard == Framework("ISO 27001") && strings.HasPrefix(isoDomain.Title, "A") {
		return nil
	}

	filename := fmt.Sprintf("pkg/mapper/%s/%s.md", safeFileName(string(standard)), safeFileName(shortenDomain(standard, isoDomain.Title)))

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	doc := md.NewMarkdown(f).
		H1(fmt.Sprintf("%s - %s", standard, string(isoDomain.Title)))

	for _, isoControl := range isoDomain.Controls {
		doc.H2(isoControl.Ref).
			PlainText(md.Bold(isoControl.Title) + "\n").
			PlainText(isoControl.Summary).
			LF()

		fcids := []string{}

		for scfID, controlMapping := range scfControlMapping {
			soc2FrameworkControlIDs := controlMapping[standard]
			for _, fcid := range soc2FrameworkControlIDs {
				if FCIDToAnnex(standard, string(fcid)) == isoControl.Ref {
					link := fmt.Sprintf("[%s](../scf/%s.md)", string(scfID), safeFileName(string(scfID)))

					found := false

					for _, f := range fcids {
						if f == link {
							found = true
						}
					}

					if !found {
						fcids = append(fcids, link)
					}
				}
			}
		}

		if len(fcids) > 0 {
			slices.Sort(fcids)
			doc.H3("Mapped SCF controls").
				BulletList(fcids...).
				LF()
		}
	}

	if err := doc.Build(); err != nil {
		return err
	}

	err = generateMetadata(filename, string(standard), isoDomain.Title, isoDomain.Title, "")
	if err != nil {
		return err
	}

	return nil
}

func FCIDToAnnex(framework Framework, fcid string) string {
	if framework == Framework("ISO 27002") {
		fcid = "A" + fcid
	}

	subSubRegexPattern := `^A([0-9]+)\.([0-9]+)\.([0-9]+).*`
	subSubRegex := regexp.MustCompile(subSubRegexPattern)
	subsubAnnexMatches := subSubRegex.FindStringSubmatch(fcid)

	if len(subsubAnnexMatches) > 0 {
		return fmt.Sprintf("A.%s.%s.%s", subsubAnnexMatches[1], subsubAnnexMatches[2], subsubAnnexMatches[3])
	}

	subRegexPattern := `^A([0-9]+)\.([0-9]+).*`
	subRegex := regexp.MustCompile(subRegexPattern)
	subAnnexMatches := subRegex.FindStringSubmatch(fcid)

	if len(subAnnexMatches) > 0 {
		return fmt.Sprintf("A.%s.%s", subAnnexMatches[1], subAnnexMatches[2])
	}

	if framework == Framework("ISO 27002") {
		log.Fatal("ISO 27002 not annex")
	}

	return strings.ReplaceAll(strings.ReplaceAll(fcid, ")", ""), "(", ".")
}

func FCIDToRequirement(fcid string) string {
	return strings.ReplaceAll(strings.ReplaceAll(fcid, ")", ""), "(", ".")
}

func GenerateISOIndex(standard Framework, isoFramework ISOFramework) error {
	f, err := os.Create(fmt.Sprintf("pkg/mapper/%s/index.md", safeFileName(string(standard))))
	if err != nil {
		return err
	}

	doc := md.NewMarkdown(f).
		H1(string(standard))
	controlLinks := []string{}

	for _, domain := range isoFramework.Domains {
		controlLinks = append(controlLinks, fmt.Sprintf("[%s](%s.md)", domain.Title, safeFileName(shortenDomain(standard, domain.Title))))
	}

	doc.BulletList(controlLinks...)

	if err := doc.Build(); err != nil {
		return err
	}

	return nil
}

func shortenDomain(standard Framework, domain string) string {
	if standard == Framework("ISO 27001") {
		parts := strings.Split(domain, " -")
		return strings.ReplaceAll(parts[0], ".", "-")
	} else {
		return strings.ReplaceAll(domain[0:3], ".", "-")
	}
}
