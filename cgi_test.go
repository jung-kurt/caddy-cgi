package cgi

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// handleGet returns a cgi handler (which implements ServeHTTP) based on the
// specified directive string. The directive can contain more than one block,
// but only the first is associated with the returned handler.
func handlerGet(directiveStr, rootStr string) (hnd handlerType, err error) {
	var c *caddy.Controller
	var mids []httpserver.Middleware
	var cfg *httpserver.SiteConfig
	var ok bool
	var midLen int

	c = caddy.NewTestController("http", directiveStr)
	cfg = httpserver.GetConfig(c)
	cfg.Root = rootStr
	err = configureServer(c, cfg)
	if err == nil {
		mids = cfg.Middleware()
		midLen = len(mids)
		if midLen > 0 {
			hnd, ok = mids[0](httpserver.EmptyNext).(handlerType)
			if ok {
				// .. success
			} else {
				err = fmt.Errorf("expected middleware handler to be CGI handler")
			}
		} else {
			err = fmt.Errorf("no middlewares present")
		}
	}
	return
}

func TestServe(t *testing.T) {
	var err error
	var code int
	var hnd handlerType
	var srv *httptest.Server
	directiveList := []string{
		`cgi /servertime {.}/test/example`,
		`cgi {
  match /servertime
  except /servertime/1934
  exec {.}/test/example --example
  env CGI_GLOBAL=12
  empty_env CGI_LOCAL
}`,
	}

	requestList := []string{
		"/servertime",
		"/servertime/1930/05/11?name=Edsger%20W.%20Dijkstra",
		"/servertime/1934/02/15?name=Niklaus%20Wirth",
		"/example.txt",
	}

	expectStr := `=== Directive 0 ===
cgi /servertime {.}/test/example
--- Request /servertime ---
code [0], error [example error message]
PATH_INFO []
CGI_GLOBAL []
Arg 1 []
QUERY_STRING []
HTTP_TOKEN_CLAIM_USER [quixote]
CGI_LOCAL is unset
--- Request /servertime/1930/05/11?name=Edsger%20W.%20Dijkstra ---
code [0], error [example error message]
PATH_INFO [/1930/05/11]
CGI_GLOBAL []
Arg 1 []
QUERY_STRING [name=Edsger%20W.%20Dijkstra]
HTTP_TOKEN_CLAIM_USER [quixote]
CGI_LOCAL is unset
--- Request /servertime/1934/02/15?name=Niklaus%20Wirth ---
code [0], error [example error message]
PATH_INFO [/1934/02/15]
CGI_GLOBAL []
Arg 1 []
QUERY_STRING [name=Niklaus%20Wirth]
HTTP_TOKEN_CLAIM_USER [quixote]
CGI_LOCAL is unset
--- Request /example.txt ---
=== Directive 1 ===
cgi {
  match /servertime
  except /servertime/1934
  exec {.}/test/example --example
  env CGI_GLOBAL=12
  empty_env CGI_LOCAL
}
--- Request /servertime ---
code [0], error [example error message]
PATH_INFO []
CGI_GLOBAL [12]
Arg 1 [--example]
QUERY_STRING []
HTTP_TOKEN_CLAIM_USER [quixote]
CGI_LOCAL is set to []
--- Request /servertime/1930/05/11?name=Edsger%20W.%20Dijkstra ---
code [0], error [example error message]
PATH_INFO [/1930/05/11]
CGI_GLOBAL [12]
Arg 1 [--example]
QUERY_STRING [name=Edsger%20W.%20Dijkstra]
HTTP_TOKEN_CLAIM_USER [quixote]
CGI_LOCAL is set to []
--- Request /servertime/1934/02/15?name=Niklaus%20Wirth ---
--- Request /example.txt ---
`
	// Testing the ServeHTTP method requires OS-specific CGI scripts, because a
	// system call is made to respond to the request.
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		var buf bytes.Buffer
		for dirJ := 0; dirJ < len(directiveList) && err == nil; dirJ++ {
			hnd, err = handlerGet(directiveList[dirJ], "./test")
			if err == nil {
				srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					r.Header.Set("Token-Claim-User", "quixote")
					r.Header.Set("Token-Claim-Language", "en-GB")
					r.Header.Add("Token-Claim-Language", "en-AU")
					code, err = hnd.ServeHTTP(w, r)
					if err != nil {
						fmt.Fprintf(&buf, "code [%d], error [%s]\n", code, err)
					}
				}))
				fmt.Fprintf(&buf, "=== Directive %d ===\n%s\n", dirJ, directiveList[dirJ])
				for reqJ := 0; reqJ < len(requestList) && err == nil; reqJ++ {
					var res *http.Response
					fmt.Fprintf(&buf, "--- Request %s ---\n", requestList[reqJ])
					res, err = http.Get(srv.URL + requestList[reqJ])
					if err == nil {
						_, err = buf.ReadFrom(res.Body)
						res.Body.Close()
					}
				}
				srv.Close()
			}
		}
		if err == nil {
			gotStr := buf.String()
			if expectStr != gotStr {
				err = errorf("expected %s, got %s\n", expectStr, gotStr)
			}
		}
		if err != nil {
			t.Fatalf("%s", err)
		}
	}
}

func TestMatches(t *testing.T) {
	var ok bool
	var prefixStr, suffixStr string
	// [request, pattern, expected success:0/expected error:1, prefix, suffix]
	list := [][]string{
		{"/foo/bar/baz", "/foo", "1", "/foo", "/bar/baz"},
		{"/foo/bar/baz", "/foo/*/baz", "1", "/foo/bar/baz", ""},
		{"/foo/bar/baz", "/foo/bar", "1", "/foo/bar", "/baz"},
		{"/foo/bar/baz", "foo/bar", "0", "", ""},
	}

	for _, rec := range list {
		ok, prefixStr, suffixStr = match(rec[0], rec[1])
		if ok {
			if rec[2] != "1" || rec[3] != prefixStr || rec[4] != suffixStr {
				t.Fatalf("expected mismatch for \"%s\" and \"%s\"", rec[0], rec[1])
			}
		} else {
			if rec[2] != "0" {
				t.Fatalf("expected match for \"%s\" and \"%s\"", rec[0], rec[1])
			}
		}
	}
}
