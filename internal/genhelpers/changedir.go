package genhelpers

import (
	"os"

	"github.com/rs/zerolog/log"
)

// ChangeToRootDir changes the working directory to the root of the repository
func ChangeToRootDir(rootDir string) {
	if err := os.Chdir(rootDir); err != nil {
		log.Fatal().Err(err).Msg("failed to change working directory")
	}

	_, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get current working directory")
	}
}
