package main

import (
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/sleuth/spider"
	"github.com/theopenlane/core/pkg/sleuth/tech"
)

func main() {
	links := spider.PerformWebSpider([]string{"https://console.theopenlane.io"})
	combinedAppInfo := make(map[string]tech.AppInfo)

	for _, link := range links.Links {
		client, err := tech.NewTech(link.Link)
		if err != nil {
			log.Error().Err(err).Msgf("Error creating Tech client for link: %s", link.Link)
			continue
		}

		appInfo, err := client.GetTech()
		if err != nil {
			log.Error().Err(err).Msgf("Error fetching technology information for link: %s", link.Link)
			continue
		}

		for name, info := range appInfo {
			combinedAppInfo[name] = info
		}
	}

	tech.PrintAppInfoTable(combinedAppInfo)
}
