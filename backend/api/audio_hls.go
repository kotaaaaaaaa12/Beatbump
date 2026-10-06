package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

const audioPrefixLimit = 256 << 10
const audioIndexLimit = 64
const audioIndexTTL = 10 * time.Minute

type audioSegment struct {
	offset, size int64
	duration     float64
}
type audioIndex struct {
	initEnd, total int64
	segments       []audioSegment
	prefix         []byte
}
type audioIndexEntry struct {
	ready   chan struct{}
	expires time.Time
	index   *audioIndex
	err     error
}
type audioIndexMemory struct {
	sync.Mutex
	entries map[[32]byte]*audioIndexEntry
}

var audioIndexes = audioIndexMemory{entries: make(map[[32]byte]*audioIndexEntry)}

// Parse the bounded initialization prefix and its flat Segment Index box.
// Offsets remain in the original file: audio is never transcoded or reordered.
func parseAudioIndex(prefix []byte, total int64) (*audioIndex, error) {
	invalid := errors.New("Unsupported audio segment index")
	if len(prefix) < 24 || total < int64(len(prefix)) || string(prefix[4:8]) != "ftyp" {
		return nil, invalid
	}
	var initEnd int64
	for pos := 0; pos+8 <= len(prefix); {
		size := int64(binary.BigEndian.Uint32(prefix[pos : pos+4]))
		header := int64(8)
		if size == 1 {
			if pos+16 > len(prefix) {
				return nil, invalid
			}
			raw := binary.BigEndian.Uint64(prefix[pos+8 : pos+16])
			if raw > math.MaxInt64 {
				return nil, invalid
			}
			size = int64(raw)
			header = 16
		}
		if size < header || size > int64(len(prefix)-pos) {
			return nil, invalid
		}
		end := pos + int(size)
		kind := string(prefix[pos+4 : pos+8])
		if kind == "moov" {
			if initEnd != 0 || !bytes.Contains(prefix[pos:end], []byte("mvex")) {
				return nil, invalid
			}
			initEnd = int64(end)
		}
		if kind == "sidx" {
			if initEnd == 0 {
				return nil, invalid
			}
			data := prefix[pos+int(header) : end]
			if len(data) < 24 || data[0] > 1 {
				return nil, invalid
			}
			timescale := binary.BigEndian.Uint32(data[8:12])
			if timescale == 0 {
				return nil, invalid
			}
			p := 12
			var firstOffset uint64
			if data[0] == 0 {
				firstOffset = uint64(binary.BigEndian.Uint32(data[p+4 : p+8]))
				p += 8
			} else {
				if len(data) < 32 {
					return nil, invalid
				}
				firstOffset = binary.BigEndian.Uint64(data[p+8 : p+16])
				p += 16
			}
			count := int(binary.BigEndian.Uint16(data[p+2 : p+4]))
			p += 4
			if count == 0 || count > 2048 || p+count*12 != len(data) || firstOffset > uint64(total) {
				return nil, invalid
			}
			offset := int64(end) + int64(firstOffset)
			if offset < initEnd || offset >= total {
				return nil, invalid
			}
			result := &audioIndex{initEnd: initEnd, total: total, prefix: prefix}
			for i := 0; i < count; i++ {
				reference := binary.BigEndian.Uint32(data[p : p+4])
				duration := binary.BigEndian.Uint32(data[p+4 : p+8])
				sap := binary.BigEndian.Uint32(data[p+8 : p+12])
				p += 12
				size := int64(reference & 0x7fffffff)
				seconds := float64(duration) / float64(timescale)
				if reference>>31 != 0 || size < 8 || size > 5<<20 || duration == 0 || seconds > 60 || sap>>31 != 1 || (sap>>28)&7 != 1 || size > total-offset {
					return nil, invalid
				}
				result.segments = append(result.segments, audioSegment{offset, size, seconds})
				offset += size
			}
			if offset != total {
				return nil, invalid
			}
			return result, nil
		}
		if kind == "moof" || kind == "mdat" {
			return nil, invalid
		}
		pos = end
	}
	return nil, invalid
}

func (memory *audioIndexMemory) lookup(key [32]byte) *audioIndexEntry {
	now := time.Now()
	memory.Lock()
	defer memory.Unlock()
	entry := memory.entries[key]
	if entry != nil && !now.Before(entry.expires) {
		delete(memory.entries, key)
		return nil
	}
	return entry
}

func (memory *audioIndexMemory) load(ctx context.Context, u *url.URL) (*audioIndex, error) {
	key := sha256.Sum256([]byte(u.String()))
	now := time.Now()
	memory.Lock()
	for k, e := range memory.entries {
		if !now.Before(e.expires) {
			delete(memory.entries, k)
		}
	}
	entry := memory.entries[key]
	if entry != nil {
		memory.Unlock()
		select {
		case <-entry.ready:
			return entry.index, entry.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if len(memory.entries) >= audioIndexLimit {
		var oldest [32]byte
		var expiry time.Time
		for k, e := range memory.entries {
			select {
			case <-e.ready:
				if expiry.IsZero() || e.expires.Before(expiry) {
					oldest, expiry = k, e.expires
				}
			default:
			}
		}
		if expiry.IsZero() {
			memory.Unlock()
			return nil, errors.New("Audio startup is busy")
		}
		delete(memory.entries, oldest)
	}
	entry = &audioIndexEntry{ready: make(chan struct{}), expires: now.Add(30 * time.Second)}
	memory.entries[key] = entry
	memory.Unlock()
	index, err := loadAudioIndex(ctx, u)
	memory.Lock()
	entry.index, entry.err = index, err
	if err != nil {
		delete(memory.entries, key)
	} else {
		entry.expires = time.Now().Add(audioIndexTTL)
	}
	close(entry.ready)
	memory.Unlock()
	return index, err
}

func loadAudioIndex(ctx context.Context, u *url.URL) (*audioIndex, error) {
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	companion, direct, err := mediaSourceRequests(ctx, u, fmt.Sprintf("bytes=0-%d", audioPrefixLimit-1))
	if err != nil {
		return nil, err
	}
	response, _, _, err := fetchMedia(ctx, companion, direct, true, mediaHedgeDelay)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	parts := contentRangePattern.FindStringSubmatch(response.Header.Get("Content-Range"))
	if len(parts) != 4 {
		return nil, errors.New("Missing audio range")
	}
	total, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, audioPrefixLimit+1))
	if err != nil {
		return nil, err
	}
	want := int64(audioPrefixLimit)
	if total < want {
		want = total
	}
	if int64(len(data)) != want || response.ContentLength != want {
		return nil, errors.New("Incomplete audio initialization")
	}
	return parseAudioIndex(data, total)
}

// Serve the initialization and complete first segments from the bounded prefix.
// All other ranges keep the existing streamed transport and validation.
func (memory *audioIndexMemory) prefixResponse(source, requestedRange string) *http.Response {
	if requestedRange == "" {
		return nil
	}
	entry := memory.lookup(sha256.Sum256([]byte(source)))
	if entry == nil {
		return nil
	}
	select {
	case <-entry.ready:
	default:
		return nil
	}
	if entry.err != nil || entry.index == nil {
		return nil
	}
	index := entry.index
	requested := strings.Split(strings.TrimPrefix(requestedRange, "bytes="), "-")
	if len(requested) != 2 {
		return nil
	}
	start, e1 := strconv.ParseInt(requested[0], 10, 64)
	var end int64
	var e2 error
	if requested[1] == "" {
		end = index.total - 1
	} else {
		end, e2 = strconv.ParseInt(requested[1], 10, 64)
	}
	if e1 != nil || e2 != nil || start < 0 || end < start || end >= int64(len(index.prefix)) {
		return nil
	}
	size := end - start + 1
	return &http.Response{StatusCode: 206, ContentLength: size, Header: http.Header{
		"Content-Type": []string{"audio/mp4"}, "Content-Length": []string{strconv.FormatInt(size, 10)},
		"Content-Range": []string{fmt.Sprintf("bytes %d-%d/%d", start, end, index.total)}, "Accept-Ranges": []string{"bytes"},
	}, Body: io.NopCloser(bytes.NewReader(index.prefix[start : end+1]))}
}

func audioPlaylist(index *audioIndex, mediaURI string) string {
	longest := 0.0
	for _, segment := range index.segments {
		longest = math.Max(longest, segment.duration)
	}
	var out strings.Builder
	fmt.Fprintf(&out, "#EXTM3U\n#EXT-X-VERSION:7\n#EXT-X-TARGETDURATION:%d\n#EXT-X-MEDIA-SEQUENCE:0\n#EXT-X-PLAYLIST-TYPE:VOD\n#EXT-X-INDEPENDENT-SEGMENTS\n#EXT-X-MAP:URI=\"%s\",BYTERANGE=\"%d@0\"\n", int(math.Ceil(longest)), mediaURI, index.initEnd)
	for _, segment := range index.segments {
		fmt.Fprintf(&out, "#EXTINF:%.6f,\n#EXT-X-BYTERANGE:%d@%d\n%s\n", segment.duration, segment.size, segment.offset, mediaURI)
	}
	out.WriteString("#EXT-X-ENDLIST\n")
	return out.String()
}

func CloudAudioManifestHandler(c echo.Context) error {
	u, err := mediaTicketSource(c)
	if err != nil {
		return err
	}
	if u.Query().Get("itag") != "140" {
		return c.String(422, "Unsupported audio stream")
	}
	index, err := audioIndexes.load(c.Request().Context(), u)
	if err != nil {
		return c.String(502, "Audio segments unavailable. Use regular playback.")
	}
	playlist := audioPlaylist(index, "/api/v1/media?ticket="+url.QueryEscape(c.QueryParam("ticket")))
	c.Response().Header().Set("Cache-Control", "no-store")
	c.Response().Header().Set("X-Beatbump-Audio-Revision", "native-hls-1")
	c.Response().Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	c.Response().Header().Set("Content-Length", strconv.Itoa(len(playlist)))
	if c.Request().Method == http.MethodHead {
		c.Response().WriteHeader(200)
		return nil
	}
	return c.String(200, playlist)
}
