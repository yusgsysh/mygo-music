package music

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// A cover's URL can name the cookie to fetch it with in its fragment, which
// goes as a Cookie header and never reaches the server as part of the address.
func TestFetchCoverCookie(t *testing.T) {
	var cookie, query, fragment string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, query, fragment = r.Header.Get("Cookie"), r.URL.RawQuery, r.URL.Fragment
		_, _ = w.Write([]byte("cover"))
	}))
	defer srv.Close()

	u := srv.URL + "/cover?coverId=x#mygo-cookie=" + url.QueryEscape("music-token=abc")
	got, err := Fetch(u)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "cover" {
		t.Errorf("got %q", got)
	}
	if cookie != "music-token=abc" {
		t.Errorf("cookie: %q", cookie)
	}
	if query != "coverId=x" || fragment != "" {
		t.Errorf("query: %q, fragment: %q", query, fragment)
	}
}
