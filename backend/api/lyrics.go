package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/labstack/echo/v4"
)

// Lyrics use a fixed upstream; no caller-supplied host or credentials are forwarded.
var lyricsHTTP = &http.Client{Transport: proxyHTTP.Transport, Timeout: 8 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 || req.URL.Scheme != "https" || req.URL.Host != "lrclib.net" || req.URL.User != nil {
			return errors.New("Invalid lyrics redirect")
		}
		return nil
	},
}
var lyricsSlots = make(chan struct{}, 4)
var lyricsMemory = struct {
	sync.Mutex
	entries map[string]lyricsCacheEntry
}{entries: make(map[string]lyricsCacheEntry)}

type lyricsCacheEntry struct {
	result  lyricsResult
	expires time.Time
}
type lyricsRecord struct {
	ID           int64   `json:"id"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}
type lyricsResult struct {
	Status string        `json:"status"`
	Source string        `json:"source,omitempty"`
	Record *lyricsRecord `json:"record,omitempty"`
}

var lyricsDecoration = regexp.MustCompile(`(?i)\s*[\[(](?:official\s*)?(?:audio|video|music video|lyric(?:s| video)?|visualizer)[\])]\s*`)

func cleanLyricsTitle(s string) string {
	return strings.TrimSpace(lyricsDecoration.ReplaceAllString(s, " "))
}
func lyricsName(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, s)
}
func matchingLyrics(record lyricsRecord, title, artist string, duration float64) bool {
	if lyricsName(record.TrackName) != lyricsName(title) || lyricsName(record.ArtistName) != lyricsName(artist) {
		return false
	}
	if math.IsNaN(record.Duration) || math.IsInf(record.Duration, 0) || record.Duration <= 0 || math.Abs(record.Duration-duration) > 2 {
		return false
	}
	return record.ID > 0 && len(record.PlainLyrics) <= 65536 && len(record.SyncedLyrics) <= 65536 && (record.Instrumental || strings.TrimSpace(record.PlainLyrics) != "" || strings.TrimSpace(record.SyncedLyrics) != "")
}
func requestLyrics(ctx context.Context, path string, query url.Values, dest any) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://lrclib.net/api/"+path+"?"+query.Encode(), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "BeatBump/2.2 (https://github.com/giwty/Beatbump)")
	req.Header.Set("Accept", "application/json")
	response, err := lyricsHTTP.Do(req)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if response.StatusCode != http.StatusOK {
		return false, errors.New("Lyrics service unavailable")
	}
	const limit = 3 * 1024 * 1024
	body, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil || len(body) > limit {
		return false, errors.New("Invalid lyrics response")
	}
	if err = json.Unmarshal(body, dest); err != nil {
		return false, err
	}
	return true, nil
}
func resolveLyrics(ctx context.Context, title, artist, album string, duration float64) (lyricsResult, error) {
	query := url.Values{"track_name": {title}, "artist_name": {artist}, "duration": {strconv.FormatFloat(duration, 'f', 3, 64)}}
	if album != "" {
		query.Set("album_name", album)
	}
	var exact lyricsRecord
	found, err := requestLyrics(ctx, "get", query, &exact)
	if err != nil {
		return lyricsResult{}, err
	}
	if found && matchingLyrics(exact, title, artist, duration) {
		return lyricsResult{Status: "found", Source: "LRCLIB", Record: &exact}, nil
	}
	// Search only after a genuine miss. Never turn a provider failure into "no lyrics".
	query.Del("duration")
	query.Del("album_name")
	var records []lyricsRecord
	_, err = requestLyrics(ctx, "search", query, &records)
	if err != nil {
		return lyricsResult{}, err
	}
	var best *lyricsRecord
	for i := range records {
		r := &records[i]
		if !matchingLyrics(*r, title, artist, duration) {
			continue
		}
		if best == nil || (best.SyncedLyrics == "" && r.SyncedLyrics != "") || ((best.SyncedLyrics == "") == (r.SyncedLyrics == "") && math.Abs(r.Duration-duration) < math.Abs(best.Duration-duration)) {
			best = r
		}
	}
	if best != nil {
		return lyricsResult{Status: "found", Source: "LRCLIB", Record: best}, nil
	}
	return lyricsResult{Status: "not_found"}, nil
}
func LyricsEndpointHandler(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	q := c.QueryParams()
	title := cleanLyricsTitle(q.Get("title"))
	artist := strings.TrimSpace(q.Get("artist"))
	album := strings.TrimSpace(q.Get("album"))
	duration, err := strconv.ParseFloat(q.Get("duration"), 64)
	if err != nil || math.IsNaN(duration) || math.IsInf(duration, 0) || duration < 1 || duration > 3600 || title == "" || artist == "" || len(title) > 300 || len(artist) > 300 || len(album) > 300 {
		return c.JSON(http.StatusBadRequest, lyricsResult{Status: "invalid_track"})
	}
	key := url.Values{"title": {title}, "artist": {artist}, "album": {album}, "duration": {strconv.FormatFloat(duration, 'f', 3, 64)}}.Encode()
	lyricsMemory.Lock()
	cached, ok := lyricsMemory.entries[key]
	lyricsMemory.Unlock()
	if ok && time.Now().Before(cached.expires) {
		return c.JSON(http.StatusOK, cached.result)
	}
	select {
	case lyricsSlots <- struct{}{}:
		defer func() { <-lyricsSlots }()
	default:
		return c.JSON(http.StatusServiceUnavailable, lyricsResult{Status: "unavailable"})
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 12*time.Second)
	defer cancel()
	result, err := resolveLyrics(ctx, title, artist, album, duration)
	if err != nil {
		return c.JSON(http.StatusBadGateway, lyricsResult{Status: "unavailable"})
	}
	ttl := 6 * time.Hour
	if result.Status == "not_found" {
		ttl = 5 * time.Minute
	}
	lyricsMemory.Lock()
	for k, v := range lyricsMemory.entries {
		if time.Now().After(v.expires) {
			delete(lyricsMemory.entries, k)
		}
	}
	if len(lyricsMemory.entries) >= 128 {
		for k := range lyricsMemory.entries {
			delete(lyricsMemory.entries, k)
			break
		}
	}
	lyricsMemory.entries[key] = lyricsCacheEntry{result: result, expires: time.Now().Add(ttl)}
	lyricsMemory.Unlock()
	return c.JSON(http.StatusOK, result)
}
