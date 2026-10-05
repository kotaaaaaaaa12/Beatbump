package api

import (
	"beatbump-server/backend/_youtube"
	youtubeapi "beatbump-server/backend/_youtube/api"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync"
	"time"
)

// Country-selected charts are independent of the Container's egress location.
// Never label an unverified, IP-selected home shelf as Japanese content.
const japanRegionRevision = "country-selected-charts-1"

var japanChartCache struct {
	sync.Mutex
	shelves []Carousel
	expires time.Time
	pending chan struct{}
	err     error
}

func japanCharts() ([]Carousel, error) {
	for {
		japanChartCache.Lock()
		if time.Now().Before(japanChartCache.expires) {
			shelves, err := japanChartCache.shelves, japanChartCache.err
			japanChartCache.Unlock()
			return shelves, err
		}
		if pending := japanChartCache.pending; pending != nil {
			// Serve previously verified JP content while it refreshes.
			if len(japanChartCache.shelves) != 0 {
				shelves := japanChartCache.shelves
				japanChartCache.Unlock()
				return shelves, nil
			}
			japanChartCache.Unlock()
			<-pending
			continue
		}
		japanChartCache.pending = make(chan struct{})
		stale := japanChartCache.shelves
		japanChartCache.Unlock()
		if len(stale) != 0 {
			go refreshJapanCharts()
			return stale, nil
		}
		refreshJapanCharts()
	}
}

func refreshJapanCharts() {
	raw, err := youtubeapi.Browse("FEmusic_charts", youtubeapi.PageType_MusicPageTypePlaylist, "", nil, nil, nil, youtubeapi.WebMusic)
	var shelves []Carousel
	if err == nil {
		shelves, err = parseJapanCharts(raw)
	}
	japanChartCache.Lock()
	defer japanChartCache.Unlock()
	if err == nil {
		japanChartCache.shelves = shelves
		japanChartCache.err = nil
		japanChartCache.expires = time.Now().Add(10 * time.Minute)
	} else {
		log.Printf("Japan charts unavailable: %v", err)
		japanChartCache.err = err
		// Retry later; retain only data previously verified as Japanese.
		if len(japanChartCache.shelves) != 0 {
			japanChartCache.err = nil
		}
		japanChartCache.expires = time.Now().Add(30 * time.Second)
	}
	close(japanChartCache.pending)
	japanChartCache.pending = nil
}

func parseJapanCharts(raw []byte) ([]Carousel, error) {
	// Verify the selected dropdown title, not an unselected Japan menu option.
	var selection struct {
		Contents struct {
			SingleColumn struct {
				Tabs []struct {
					Tab struct {
						Content struct {
							Sections struct {
								Contents []struct {
									Shelf struct {
										Subheaders []struct {
											Aligned struct {
												Items []struct {
													Button struct {
														Title struct {
															Runs []struct {
																Text string `json:"text"`
															} `json:"runs"`
														} `json:"title"`
													} `json:"musicSortFilterButtonRenderer"`
												} `json:"startItems"`
											} `json:"musicSideAlignedItemRenderer"`
										} `json:"subheaders"`
									} `json:"musicShelfRenderer"`
								} `json:"contents"`
							} `json:"sectionListRenderer"`
						} `json:"content"`
					} `json:"tabRenderer"`
				} `json:"tabs"`
			} `json:"singleColumnBrowseResultsRenderer"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(raw, &selection); err != nil {
		return nil, err
	}
	verified := false
	for _, tab := range selection.Contents.SingleColumn.Tabs {
		for _, section := range tab.Tab.Content.Sections.Contents {
			for _, header := range section.Shelf.Subheaders {
				for _, item := range header.Aligned.Items {
					if len(item.Button.Title.Runs) != 0 && item.Button.Title.Runs[0].Text == "Japan" {
						verified = true
					}
				}
			}
		}
	}
	if !verified {
		return nil, errors.New("upstream did not select Japan")
	}
	var response _youtube.HomeResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, err
	}
	parsed := ParseHome(response).(map[string]interface{})
	var shelves []Carousel
	for _, shelf := range parsed["carousels"].([]Carousel) {
		switch shelf.Header.Title {
		case "Video charts":
			shelf.Header.Title = "Japan charts"
			// A mismatched regional playlist is never accepted as a JP fallback.
			for i, item := range shelf.Contents {
				if !strings.Contains(item.Title, "Japan") {
					return nil, errors.New("upstream returned a non-Japan chart")
				}
				switch item.Title {
				case "Top 100 Live Performances - Japan", "Trending 20 Japan", "Daily Top Music Videos - Japan", "Top 100 Music Videos Japan":
					shelf.Contents[i].TranslationKey = item.Title
				}
			}
		case "Top artists":
			shelf.Header.Title = "Popular artists in Japan"
		default:
			continue
		}
		shelf.Header.BrowseId = nil
		shelf.Header.Params = nil
		shelves = append(shelves, shelf)
	}
	if len(shelves) == 0 {
		return nil, errors.New("Japan charts contained no shelves")
	}
	return shelves, nil
}

func marketSelectedShelf(title string) bool {
	title = strings.ToLower(strings.ReplaceAll(title, "’", "'"))
	return strings.Contains(title, "community playlists") ||
		strings.HasPrefix(title, "shorts featured section") ||
		title == "featured playlists for you" || title == "featured playlists" ||
		title == "today's biggest hits" || title == "quick picks" || title == "trending"
}

func marketSelectedPlaylist(title string) bool {
	title = strings.ToLower(strings.ReplaceAll(title, "’", "'"))
	return title == "today's hits" || title == "new & now" || title == "new and now" ||
		strings.HasPrefix(title, "top 100 music videos") ||
		strings.HasPrefix(title, "daily top music videos") ||
		strings.HasPrefix(title, "trending 20 ") ||
		strings.HasPrefix(title, "top 100 live performances")
}

func applyJapanHomeRegion(response map[string]interface{}, initial bool, getCharts func() ([]Carousel, error)) {
	shelves, _ := response["carousels"].([]Carousel)
	cleaned := make([]Carousel, 0, len(shelves)+2)
	response["contentRegion"] = "JP"
	response["regionRevision"] = japanRegionRevision
	if initial {
		charts, err := getCharts()
		if err == nil && len(charts) != 0 {
			cleaned = append(cleaned, charts...)
			response["regionalStatus"] = "ready"
		} else {
			response["regionalStatus"] = "unavailable"
		}
	}
	for _, shelf := range shelves {
		if marketSelectedShelf(shelf.Header.Title) {
			continue
		}
		items := make([]IListItemRenderer, 0, len(shelf.Contents))
		for _, item := range shelf.Contents {
			if !marketSelectedPlaylist(item.Title) {
				items = append(items, item)
			}
		}
		shelf.Contents = items
		if len(items) != 0 || shelf.Categories != nil {
			cleaned = append(cleaned, shelf)
		}
	}
	response["carousels"] = cleaned
}
