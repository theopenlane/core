package jobs

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/providers/resend"
)

type EmailArgs struct {
	Message newman.EmailMessage `json:"message"`
}

func (EmailArgs) Kind() string { return "email" }

type EmailWorker struct {
	river.WorkerDefaults[EmailArgs]

	EmailConfig
}

// EmailConfig contains the configuration for the email worker
type EmailConfig struct {
	// DevMode is a flag to enable dev mode
	DevMode bool `json:"dev_mode"`
	// TestDir is the directory to use for dev mode
	TestDir string `json:"test_dir"`
	// Token is the token to use for the email provider
	Token string `json:"token"`
	// FromEmail is the email address to use as the sender
	FromEmail string `json:"from_email"`
}

// validateEmailConfig validates the email configuration settings
func (w *EmailWorker) validateEmailConfig() error {
	if w.DevMode && w.TestDir == "" {
		return fmt.Errorf("missing test directory") // nolint:goerr113
	}

	if !w.DevMode && w.Token == "" {
		return fmt.Errorf("missing token") // nolint:goerr113
	}

	return nil
}

func (w *EmailWorker) Work(ctx context.Context, job *river.Job[EmailArgs]) error {
	if err := w.validateEmailConfig(); err != nil {
		return err
	}

	log.Info().Strs("to", job.Args.Message.To).
		Str("subject", job.Args.Message.Subject).
		Msg("sending email")

	opts := []resend.Option{}

	if w.DevMode {
		log.Debug().Str("directory", w.TestDir).Msg("running in dev mode")

		opts = append(opts, resend.WithDevMode(w.TestDir))
	}

	if job.Args.Message.From == "" {
		job.Args.Message.From = w.FromEmail
	}

	client, err := resend.New(w.Token, opts...)
	if err != nil {
		return err
	}

	return client.SendEmailWithContext(ctx, &job.Args.Message)
}
