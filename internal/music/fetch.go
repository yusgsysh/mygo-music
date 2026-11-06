package music

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Agent is who the app says it is when fetching what it plays and shows.
const Agent = "MyGoMusic/1.0"

// fetchClient gets covers: they're small, and a slow one shouldn't hold a
// slot for long.
var fetchClient = &http.Client{Timeout: 30 * time.Second}

// maxFetch is the most a cover may take.
const maxFetch = 64 << 20

// Fetch returns what's at url, of the web or a file://, such as a cover. A
// cover's URL may name the cookie to ask for it with, as
// #mygo-cookie=name%3Dvalue, for a source that keeps its pictures behind its
// session, as this app asks for covers itself, without the source's headers.
func Fetch(u string) ([]byte, error) {
	if path, ok := PathOf(u); ok {
		return os.ReadFile(path)
	}
	u, cookie := coverCookie(u)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	req.Header.Set("User-Agent", Agent)
	res, err := fetchClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}
	return io.ReadAll(io.LimitReader(res.Body, maxFetch))
}

// coverCookie takes the cookie a cover's URL names in its fragment, as in
// #mygo-cookie=name%3Dvalue, out of the URL, so that it goes as a Cookie
// header rather than as part of the address. A fragment never reaches the
// server, so a URL that no one reads it from still asks for nothing special.
// It answers the URL without it, and "" when the URL names no cookie.
func coverCookie(u string) (string, string) {
	parsed, err := url.Parse(u)
	if err != nil {
		return u, ""
	}
	rest, ok := strings.CutPrefix(parsed.Fragment, "mygo-cookie=")
	if !ok {
		return u, ""
	}
	cookie, err := url.QueryUnescape(rest)
	if err != nil {
		return u, ""
	}
	parsed.Fragment = ""
	return parsed.String(), cookie
}

// FileURL is the file:// URL of a path.
func FileURL(path string) string {
	path = filepath.ToSlash(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path // C:/Music
	}
	return (&url.URL{Scheme: "file", Path: path}).String()
}

// PathOf returns the path a file:// URL names, and whether u is one.
func PathOf(u string) (string, bool) {
	if !strings.HasPrefix(u, "file:") {
		return "", false
	}
	parsed, err := url.Parse(u)
	if err != nil || parsed.Path == "" {
		return "", false
	}
	path := parsed.Path
	if runtime.GOOS == "windows" {
		path = strings.TrimPrefix(path, "/")
	}
	return filepath.FromSlash(path), true
}
