package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const mediaHedgeDelay = 300 * time.Millisecond
const mediaStartTimeout = 20 * time.Second

type mediaResponseError struct{ status int }

func (err *mediaResponseError) Error() string { return "Audio request rejected" }

var contentRangePattern = regexp.MustCompile(`^bytes ([0-9]+)-([0-9]+)/([0-9]+)$`)

// Each candidate must provide valid range headers and actual bytes, not just 200.
func validMediaResponse(response *http.Response, requestedRange string) error {
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusPartialContent {
		return &mediaResponseError{status: response.StatusCode}
	}
	mime := strings.ToLower(strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0]))
	if !strings.HasPrefix(mime, "audio/") && !strings.HasPrefix(mime, "video/") && mime != "application/octet-stream" {
		return errors.New("Invalid audio response")
	}
	if requestedRange == "" {
		return nil
	}
	parts := contentRangePattern.FindStringSubmatch(response.Header.Get("Content-Range"))
	if response.StatusCode != http.StatusPartialContent || len(parts) != 4 {
		return errors.New("Missing audio byte range")
	}
	start, e1 := strconv.ParseInt(parts[1], 10, 64)
	end, e2 := strconv.ParseInt(parts[2], 10, 64)
	total, e3 := strconv.ParseInt(parts[3], 10, 64)
	requested := strings.Split(strings.TrimPrefix(requestedRange, "bytes="), "-")
	wantStart, e4 := strconv.ParseInt(requested[0], 10, 64)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || start != wantStart || end < start || total <= end {
		return errors.New("Invalid audio byte range")
	}
	if requested[1] != "" {
		wantEnd, err := strconv.ParseInt(requested[1], 10, 64)
		if err != nil || end > wantEnd {
			return errors.New("Invalid audio byte range")
		}
	}
	if response.ContentLength >= 0 && response.ContentLength != end-start+1 {
		return errors.New("Invalid audio range length")
	}
	return nil
}

type mediaBody struct {
	io.Reader
	original io.ReadCloser
	cancel   context.CancelFunc
}

func (body *mediaBody) Close() error { defer body.cancel(); return body.original.Close() }

type mediaCandidate struct {
	response  *http.Response
	transport string
	err       error
}

// Remember the working transport for this exact signed stream, not another track.
// Entries contain only hashes and expire quickly; no audio or credentials are cached.
const mediaRouteLimit = 256
const mediaRouteTTL = 2 * time.Minute

type mediaRouteEntry struct {
	transport string
	expires   time.Time
}
type mediaRouteMemory struct {
	sync.Mutex
	entries map[[32]byte]mediaRouteEntry
}

var mediaRoutes = mediaRouteMemory{entries: make(map[[32]byte]mediaRouteEntry)}

func (memory *mediaRouteMemory) preferred(key [32]byte, now time.Time) string {
	memory.Lock()
	defer memory.Unlock()
	entry, exists := memory.entries[key]
	if exists && now.Before(entry.expires) {
		return entry.transport
	}
	delete(memory.entries, key)
	return "companion"
}
func (memory *mediaRouteMemory) remember(key [32]byte, transport string, now time.Time) {
	memory.Lock()
	defer memory.Unlock()
	for k, entry := range memory.entries {
		if !now.Before(entry.expires) {
			delete(memory.entries, k)
		}
	}
	if len(memory.entries) >= mediaRouteLimit {
		var oldest [32]byte
		var expiry time.Time
		for k, entry := range memory.entries {
			if expiry.IsZero() || entry.expires.Before(expiry) {
				oldest, expiry = k, entry.expires
			}
		}
		delete(memory.entries, oldest)
	}
	memory.entries[key] = mediaRouteEntry{transport, now.Add(mediaRouteTTL)}
}

// Start the proven route first; hedge the other route on delay or rejection.
// Only the first usable response continues streaming.
func fetchMedia(ctx context.Context, companion, direct *http.Request, probe bool, hedge time.Duration) (*http.Response, string, time.Duration, error) {
	started := time.Now()
	companionContext, cancelCompanion := context.WithCancel(ctx)
	directContext, cancelDirect := context.WithCancel(ctx)
	results := make(chan mediaCandidate)
	done := make(chan struct{})
	defer close(done)
	launch := func(client *http.Client, request *http.Request, child context.Context, cancel context.CancelFunc, transport string) {
		go func() {
			timer := time.AfterFunc(mediaStartTimeout, cancel)
			response, err := client.Do(request.WithContext(child))
			if err == nil {
				err = validMediaResponse(response, request.Header.Get("Range"))
			}
			var prefix []byte
			if err == nil && probe {
				prefix = make([]byte, 4096)
				var n int
				n, err = io.ReadAtLeast(response.Body, prefix, 1)
				prefix = prefix[:n]
			}
			timer.Stop()
			if err == nil && child.Err() != nil {
				err = child.Err()
			}
			if err != nil {
				if response != nil {
					response.Body.Close()
				}
				cancel()
				response = nil
			} else {
				response.Body = &mediaBody{Reader: io.MultiReader(bytes.NewReader(prefix), response.Body), original: response.Body, cancel: cancel}
			}
			select {
			case results <- mediaCandidate{response, transport, err}:
			case <-done:
				if response != nil {
					response.Body.Close()
				}
			}
		}()
	}
	key := sha256.Sum256([]byte(direct.URL.String()))
	preferred := mediaRoutes.preferred(key, started)
	companionStarted, directStarted := false, false
	startTransport := func(transport string) {
		if transport == "direct" && !directStarted {
			directStarted = true
			launch(directMediaHTTP, direct, directContext, cancelDirect, "direct")
		} else if transport == "companion" && !companionStarted {
			companionStarted = true
			launch(proxyHTTP, companion, companionContext, cancelCompanion, "companion")
		}
	}
	fallback := "direct"
	if preferred == "direct" {
		fallback = "companion"
	}
	startTransport(preferred)
	timer := time.NewTimer(hedge)
	defer timer.Stop()
	completed := 0
	var failure error
	for {
		select {
		case result := <-results:
			completed++
			if result.err == nil {
				if result.transport == "companion" {
					cancelDirect()
				} else {
					cancelCompanion()
				}
				mediaRoutes.remember(key, result.transport, time.Now())
				return result.response, result.transport, time.Since(started), nil
			}
			if failure == nil || result.transport == "direct" {
				failure = result.err
			}
			startTransport(fallback)
			if completed == 2 {
				cancelCompanion()
				cancelDirect()
				return nil, "", time.Since(started), failure
			}
		case <-timer.C:
			startTransport(fallback)
		case <-ctx.Done():
			cancelCompanion()
			cancelDirect()
			return nil, "", time.Since(started), ctx.Err()
		}
	}
}

type mediaFlushWriter struct {
	writer  io.Writer
	flusher http.Flusher
}

func (writer mediaFlushWriter) Write(bytes []byte) (int, error) {
	n, err := writer.writer.Write(bytes)
	if n > 0 {
		writer.flusher.Flush()
	}
	return n, err
}
func mediaTimingHeader(transport string, elapsed time.Duration) string {
	return fmt.Sprintf("media_%s;dur=%.1f", transport, float64(elapsed)/float64(time.Millisecond))
}

// Numeric response details make Safari byte probes visible without signed URLs.
func mediaResponseTiming(request *http.Request, response *http.Response) string {
	value := fmt.Sprintf(", media_status;dur=%d", response.StatusCode)
	if request.Method == http.MethodHead {
		value += ", media_head;dur=1"
	}
	if response.ContentLength >= 0 {
		value += fmt.Sprintf(", media_bytes;dur=%d", response.ContentLength)
	}
	parts := contentRangePattern.FindStringSubmatch(response.Header.Get("Content-Range"))
	if len(parts) == 4 {
		value += fmt.Sprintf(", media_range_start;dur=%s, media_range_end;dur=%s, media_total_bytes;dur=%s", parts[1], parts[2], parts[3])
	}
	return value
}
