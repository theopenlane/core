package internal

import (
	"bufio"
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
	"strconv"
	"strings"

	md "github.com/nao1215/markdown"
)

type GDPRFramework []GDPRArticle

type GDPRArticle struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Body        string           `json:"body"`
	Subarticles []GDPRSubarticle `json:"subarticles"`
}
type GDPRSubarticle struct {
	ID   string `json:"id"`
	Body string `json:"body"`
}

func GetGDPRControls(gdprurl string, getFile bool) (GDPRFramework, error) {
	gdprFramework := GDPRFramework{}

	if getFile {
		parsedURL, err := url.ParseRequestURI(gdprurl)
		if err != nil {
			return gdprFramework, fmt.Errorf("invalid link: %w", err)
		}

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, parsedURL.String(), nil)
		if err != nil {
			return gdprFramework, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return gdprFramework, fmt.Errorf("failed to get ISO link: %w", err)
		}

		scanner := bufio.NewScanner(resp.Body)
		articleID := -1
		subArticleID := -1

		for scanner.Scan() {
			line := strings.ReplaceAll(scanner.Text(), "### ", "")
			if line == "" {
				continue
			}

			regexPattern := `^Article\s([0-9]+)$`
			regex := regexp.MustCompile(regexPattern)
			articleMatches := regex.FindStringSubmatch(line)

			if len(articleMatches) > 0 {
				articleIDMatch, err := strconv.Atoi(articleMatches[1])
				if err != nil {
					log.Fatal("Bad GDPR article ID", articleMatches[1])
				}

				articleID = articleIDMatch - 1

				for len(gdprFramework) < articleID {
					gdprFramework = append(gdprFramework, GDPRArticle{
						ID:          fmt.Sprintf("Article %d", len(gdprFramework)),
						Subarticles: []GDPRSubarticle{},
					})
				}

				article := GDPRArticle{
					ID:          fmt.Sprintf("Article %d", articleID+1),
					Subarticles: []GDPRSubarticle{},
				}

				gdprFramework = append(gdprFramework, article)
				subArticleID = -1
			} else if articleID != -1 {
				article := gdprFramework[articleID]
				if article.Title == "" {
					article.Title = line
				} else {
					regexPattern := `^([0-9]+)\.`
					regex := regexp.MustCompile(regexPattern)
					subArticleMatches := regex.FindStringSubmatch(line)

					if len(subArticleMatches) > 0 {
						subArticleID = len(article.Subarticles)
						body := strings.ReplaceAll(line, fmt.Sprintf("%s. ", strconv.Itoa(subArticleID+1)), "")
						subarticle := GDPRSubarticle{
							ID:   fmt.Sprintf("%s.%d", article.ID, subArticleID+1),
							Body: body,
						}
						article.Subarticles = append(article.Subarticles, subarticle)
					} else {
						if subArticleID == -1 {
							if article.Body == "" {
								article.Body = line
							} else {
								article.Body = "\n" + line
							}
						} else {
							article.Subarticles[subArticleID].Body = article.Subarticles[subArticleID].Body + "\n" + line
						}
					}
				}

				gdprFramework[articleID] = article
			}
		}

		defer resp.Body.Close()

		file, err := json.MarshalIndent(gdprFramework, "", " ")
		if err != nil {
			return gdprFramework, err
		}

		err = os.WriteFile("pkg/mapper/gdpr.json", file, 0600) // nolint:mnd
		if err != nil {
			return gdprFramework, err
		}
	}

	gdprFile, err := os.Open("pkg/mapper/gdpr.json")
	if err != nil {
		return gdprFramework, err
	}

	defer gdprFile.Close()
	gdprBytes, err := io.ReadAll(gdprFile)

	if err != nil {
		return gdprFramework, err
	}

	if err := json.Unmarshal(gdprBytes, &gdprFramework); err != nil {
		return gdprFramework, err
	}

	return gdprFramework, nil
}

func GenerateGDPRMarkdown(gdprArticle GDPRArticle, scfControlMapping SCFControlMappings) error {
	scfArticle := strings.ReplaceAll(gdprArticle.ID, "Article", "Art")
	filename := fmt.Sprintf("pkg/mapper/gdpr/%s.md", safeFileName(strings.ReplaceAll(scfArticle, ".", "-")))
	f, err := os.Create(filename)

	if err != nil {
		return err
	}

	doc := md.NewMarkdown(f).
		H1(fmt.Sprintf("GDPR - %s", string(gdprArticle.ID))).
		H2(gdprArticle.Title).
		PlainText(gdprArticle.Body).
		LF()

	for _, gdprSubArticle := range gdprArticle.Subarticles {
		scfSubArticle := strings.ReplaceAll(string(gdprSubArticle.ID), "Article", "Art")
		doc.H2(gdprSubArticle.ID).
			PlainText(gdprSubArticle.Body).
			LF()

		fcids := []string{}

		for scfID, controlMapping := range scfControlMapping {
			soc2FrameworkControlIDs := controlMapping["GDPR"]
			for _, fcid := range soc2FrameworkControlIDs {
				if string(fcid) == scfSubArticle {
					fcids = append(fcids, fmt.Sprintf("[%s](../scf/%s.md)", string(scfID), safeFileName(string(scfID))))
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

	err = generateMetadata(filename, "GDPR", gdprArticle.ID, gdprArticle.Title, gdprArticle.Body)
	if err != nil {
		return err
	}

	return nil
}

func GenerateGDPRIndex(gdprFramework GDPRFramework) error {
	f, err := os.Create("pkg/mapper/gdpr/index.md")
	if err != nil {
		return err
	}

	doc := md.NewMarkdown(f).
		H1("GDPR")

	controlLinks := []string{}

	for _, article := range gdprFramework {
		if article.Title != "" {
			link := strings.ReplaceAll(article.ID, "Article", "Art")
			controlLinks = append(controlLinks, fmt.Sprintf("[%s - %s](%s.md)", article.ID, article.Title, safeFileName(link)))
		}
	}

	doc.BulletList(controlLinks...)

	if err := doc.Build(); err != nil {
		return err
	}

	return nil
}
