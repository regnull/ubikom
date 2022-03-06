package main

import (
	"fmt"
	"os"

	"github.com/regnull/ubikom/newscache"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05", NoColor: true})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/Users/regnull/gcloud/clear-talent-299521-9a3e9ed59bf1.json")

	cache := newscache.New()
	err := cache.Refresh()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to refresh news")
	}
	// Second refresh, will skip stuff.
	err = cache.Refresh()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to refresh news")
	}
	headlines := cache.GetHeadlines()
	for _, h := range headlines {
		fmt.Printf("[%d] %s\n", h.ID, h.Title)
	}

	headline, text, err := cache.GetArticle(1003)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get article")
	}
	fmt.Printf("%s\n%s\n", headline, text)
	headline, text, _ = cache.GetArticle(1003)
	headline, text, _ = cache.GetArticle(1003)
}
