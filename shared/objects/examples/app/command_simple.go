//go:build examples

package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/theopenlane/shared/objects/storage"
	"github.com/theopenlane/shared/objects/storage/providers/disk"
)

func simpleCommand() *cli.Command {
	return &cli.Command{
		Name:  "simple",
		Usage: "Run the local disk object storage example",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dir",
				Usage: "Directory to use for disk storage",
				Value: "./tmp/storage",
			},
			&cli.StringFlag{
				Name:  "local-url",
				Usage: "Local URL used when building presigned links",
				Value: "http://localhost:17608/v1/files",
			},
			&cli.BoolFlag{
				Name:  "keep",
				Usage: "Keep the storage directory after the example finishes",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg := simpleConfig{
				Dir:      cmd.String("dir"),
				LocalURL: cmd.String("local-url"),
				Keep:     cmd.Bool("keep"),
			}
			out := cmd.Writer
			if out == nil {
				out = os.Stdout
			}
			return runSimple(ctx, out, cfg)
		},
	}
}

type simpleConfig struct {
	Dir      string
	LocalURL string
	Keep     bool
}

func runSimple(ctx context.Context, out io.Writer, cfg simpleConfig) error {
	if cfg.Dir == "" {
		cfg.Dir = "./tmp/storage"
	}
	if cfg.LocalURL == "" {
		cfg.LocalURL = "http://localhost:17608/v1/files"
	}

	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		return fmt.Errorf("create storage directory: %w", err)
	}

	if !cfg.Keep {
		defer os.RemoveAll(cfg.Dir)
	}

	providerOptions := storage.NewProviderOptions(
		storage.WithBucket(cfg.Dir),
		storage.WithLocalURL(cfg.LocalURL),
	)

	provider, err := disk.NewDiskProvider(providerOptions)
	if err != nil {
		return fmt.Errorf("create disk provider: %w", err)
	}
	defer provider.Close()

	fmt.Fprintln(out, "=== Simple Object Storage Example ===")

	service := storage.NewObjectService()
	content := strings.NewReader("Hello, World! This is a test file.")

	if _, err := runLifecycle(ctx, out, service, provider, lifecycleConfig{
		FileName:        "hello.txt",
		ContentType:     "text/plain",
		Bucket:          cfg.Dir,
		Reader:          content,
		ProviderLabel:   "disk",
		PresignDuration: 15 * time.Minute,
	}); err != nil {
		return err
	}

	fmt.Fprintln(out, "\n=== Example completed successfully ===")
	return nil
}
