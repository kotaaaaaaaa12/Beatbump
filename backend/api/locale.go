package api

import (
	youtubeapi "beatbump-server/backend/_youtube/api"
	"github.com/labstack/echo/v4"
	"strings"
)

func localizedMusicClient(c echo.Context) youtubeapi.ClientInfo {
	return musicClientForLanguage(c.QueryParam("lang"))
}

func isPodcastShelf(title string) bool {
	return strings.Contains(strings.ToLower(title), "episodes") || strings.Contains(title, "エピソード")
}

func musicClientForLanguage(language string) youtubeapi.ClientInfo {
	// Copy the client: simultaneous Japanese and English requests must not race.
	client := youtubeapi.WebMusic
	if language == "ja" {
		client.Language = "ja"
	}
	return client
}
