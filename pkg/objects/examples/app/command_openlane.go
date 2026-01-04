//go:build examples

package app

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/theopenlane/core/pkg/objects/examples/config"
	"github.com/theopenlane/core/pkg/objects/examples/openlane"
)

func openlaneCommand() *cli.Command {
	return &cli.Command{
		Name:  "openlane",
		Usage: "Run the Openlane end-to-end integration example",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "api", Usage: "Openlane API base URL", Value: "http://localhost:17608"},
			&cli.StringFlag{Name: "token", Usage: "Authentication token (uses saved config if not provided)"},
			&cli.StringFlag{Name: "organization-id", Usage: "Organization ID (uses saved config if not provided)"},
			&cli.StringFlag{Name: "name", Usage: "Evidence name", Value: "Security Compliance Evidence"},
			&cli.StringFlag{Name: "description", Usage: "Evidence description", Value: "Security compliance validation evidence"},
			&cli.StringFlag{Name: "file", Usage: "Path to evidence file"},
			&cli.StringFlag{Name: "base64-output", Usage: "Path to write the file base64 response payload"},
			&cli.BoolFlag{Name: "verbose", Usage: "Enable verbose logging"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			runCfg := openlaneConfig{
				API:            cmd.String("api"),
				Token:          cmd.String("token"),
				OrganizationID: cmd.String("organization-id"),
				Name:           cmd.String("name"),
				Description:    cmd.String("description"),
				FilePath:       cmd.String("file"),
				Base64Output:   cmd.String("base64-output"),
				Verbose:        cmd.Bool("verbose"),
			}
			out := cmd.Writer
			if out == nil {
				out = os.Stdout
			}
			return runOpenlane(ctx, out, runCfg)
		},
	}
}

type openlaneConfig struct {
	API            string
	Token          string
	OrganizationID string
	Name           string
	Description    string
	FilePath       string
	Base64Output   string
	Verbose        bool
}

func runOpenlane(ctx context.Context, out io.Writer, cfg openlaneConfig) error {
	fmt.Fprintln(out, "=== Openlane End-to-End Integration Example ===")
	fmt.Fprintln(out)

	baseURL, err := url.Parse(cfg.API)
	if err != nil {
		return fmt.Errorf("invalid API URL: %w", err)
	}

	fmt.Fprintf(out, "API Endpoint: %s\n", baseURL.String())
	fmt.Fprintf(out, "Evidence Name: %s\n", cfg.Name)
	fmt.Fprintf(out, "Description: %s\n", cfg.Description)
	fmt.Fprintln(out)

	var token, orgID string

	if cfg.Token == "" || cfg.OrganizationID == "" {
		fmt.Fprintln(out, "Loading configuration from saved setup...")
		loadedCfg, err := config.Load(nil)
		if err != nil {
			return fmt.Errorf("load config (run 'setup' command first): %w", err)
		}
		token = loadedCfg.Openlane.PAT
		orgID = loadedCfg.Openlane.OrganizationID
	} else {
		token = cfg.Token
		orgID = cfg.OrganizationID
	}

	if cfg.Verbose {
		fmt.Fprintf(out, "Using organization ID: %s\n", orgID)
	}

	client, err := openlane.InitializeClient(baseURL, token, orgID)
	if err != nil {
		return fmt.Errorf("initialize client: %w", err)
	}

	evidenceFile := cfg.FilePath
	if evidenceFile == "" {
		evidenceFile = defaultEvidenceFile()
	}

	if cfg.Verbose {
		fmt.Fprintf(out, "Evidence file: %s\n", evidenceFile)
	}

	upload, err := openlane.CreateUpload(evidenceFile)
	if err != nil {
		return fmt.Errorf("load evidence file: %w", err)
	}

	fmt.Fprintln(out, "Creating evidence with file attachment...")
	evidence, err := openlane.CreateEvidenceWithFile(ctx, client, cfg.Name, cfg.Description, upload)
	if err != nil {
		return fmt.Errorf("create evidence: %w", err)
	}

	fmt.Fprintf(out, "  Evidence created: %s\n", evidence.CreateEvidence.Evidence.ID)
	fmt.Fprintf(out, "  Name: %s\n", evidence.CreateEvidence.Evidence.Name)
	if evidence.CreateEvidence.Evidence.Status != nil {
		fmt.Fprintf(out, "  Status: %s\n", *evidence.CreateEvidence.Evidence.Status)
	}

	if len(evidence.CreateEvidence.Evidence.Files.Edges) > 0 {
		fmt.Fprintf(out, "\n  Attachments (%d):\n", len(evidence.CreateEvidence.Evidence.Files.Edges))
		fileID := ""
		for _, edge := range evidence.CreateEvidence.Evidence.Files.Edges {
			if edge.Node != nil {
				fmt.Fprintf(out, "    - File ID: %s\n", edge.Node.ID)
				if fileID == "" {
					fileID = edge.Node.ID
				}
				if edge.Node.PresignedURL != nil && *edge.Node.PresignedURL != "" {
					fmt.Fprintf(out, "      Presigned URL: %s\n", *edge.Node.PresignedURL)
				}
			}
		}
	}

	fmt.Fprintln(out, "\n=== Integration test completed successfully ===")
	return nil
}

func defaultEvidenceFile() string {
	return resolvePath("testdata/sample-files/image.png")
}
