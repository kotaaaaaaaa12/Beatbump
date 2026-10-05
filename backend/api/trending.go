package api

import (
	"beatbump-server/backend/_youtube"
	"beatbump-server/backend/_youtube/api"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

func TrendingEndpointHandler(c echo.Context) error {
	browseId := c.Param("browseId")
	if browseId == "" {
		browseId = "FEmusic_explore"
	}
	qparams := c.QueryParam("params")
	client := localizedMusicClient(c)
	if browseId == "FEmusic_charts" {
		client = api.WebMusic
	}
	if browseId == "FEmusic_explore" {
		go func() { _, _ = japanCharts() }()
	}
	var responseBytes []byte
	var err error

	responseBytes, err = api.Browse(browseId, api.PageType_MusicPageTypePlaylist, qparams, nil, nil, nil, client)

	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error building API request: %s", err))
	}
	if browseId == "FEmusic_charts" {
		shelves, err := parseJapanCharts(responseBytes)
		if err != nil {
			return c.String(http.StatusBadGateway, "Japan charts unavailable")
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"carousels": shelves, "continuations": struct{}{}, "contentRegion": "JP",
			"regionRevision": japanRegionRevision, "regionalStatus": "ready",
		})
	}

	/*if category == "" {

		var exploreResponse _youtube.Explore
		err = json.Unmarshal(responseBytes, &exploreResponse)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Error building API request: %s", err))
		}

		r := parseExplore(exploreResponse)
		return c.JSON(http.StatusOK, r)
	} else {*/
	var homeResponse _youtube.HomeResponse
	err = json.Unmarshal(responseBytes, &homeResponse)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error building API request: %s", err))
	}

	r := ParseHome(homeResponse).(map[string]interface{})
	if browseId == "FEmusic_explore" {
		applyJapanHomeRegion(r, true, japanCharts)
	}

	return c.JSON(http.StatusOK, r)
	//}
}
