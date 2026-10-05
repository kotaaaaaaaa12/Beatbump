package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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

// Keep the fast companion path. Start the existing direct GET fallback early
// when companion stalls; only the first usable response continues streaming.
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
	launch(proxyHTTP, companion, companionContext, cancelCompanion, "companion")
	timer := time.NewTimer(hedge)
	defer timer.Stop()
	directStarted, completed := false, 0
	var failure error
	startDirect := func() {
		if !directStarted {
			directStarted = true
			launch(directMediaHTTP, direct, directContext, cancelDirect, "direct")
		}
	}
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
				return result.response, result.transport, time.Since(started), nil
			}
			if failure == nil || result.transport == "direct" {
				failure = result.err
			}
			startDirect()
			if completed == 2 {
				cancelCompanion()
				cancelDirect()
				return nil, "", time.Since(started), failure
			}
		case <-timer.C:
			startDirect()
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
