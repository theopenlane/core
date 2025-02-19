package main

import (
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/pkg/sleuth/sub"
)

func meow() {

	client, err := sub.NewSubfinder(sub.NewOptions())
	if err != nil {
		log.Debug().Msgf("Error creating Subfinder client: %s", err)
		panic(err)
	}

	subdomains, err := client.Enumerate()
	if err != nil {
		log.Debug().Msgf("Error enumerating subdomains: %s", err)
		panic(err)
	}

	for _, subdomain := range subdomains {
		log.Debug().Msgf("Subdomain: %s", subdomain)
	}
}

func main() {
	meow()
}
