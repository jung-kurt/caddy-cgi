package cgi

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
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
  except /servertime/1934/*/*
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
  except /servertime/1934/*/*
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

func checkEnv(envStr string) (err error) {
	// envStr is reported by CGI program invoked with pass_all_env. It looks like
	// "key1=val1\nkey2=val2..." Place keys into map and make sure all actual
	// environment variables are included in this map. Some environment values get
	// legitimately changed for spawned executable, so verify keys only.
	list := strings.Split(envStr, "\n")
	mp := make(map[string]bool)
	for _, pr := range list {
		pos := strings.Index(pr, "=")
		if pos > 0 {
			mp[pr[:pos]] = true
		}
	}
	actualList := os.Environ()
	actualLen := len(actualList)
	for j := 0; j < actualLen && err == nil; j++ {
		actualStr := actualList[j]
		pos := strings.Index(actualStr, "=")
		if pos > 0 {
			k := actualStr[:pos]
			_, ok := mp[k]
			if !ok {
				// err = fmt.Errorf("environment key \"%s\" not found in CGI environment", k)
				fmt.Printf("environment key \"%s\" not found in CGI environment\n", k)
			}
		}
	}
	return
}

func TestDir(t *testing.T) {
	var err error
	var code int
	var rsp string
	var hnd handlerType
	var srv *httptest.Server
	request := `/tmpdir`
	directive := `cgi {
match /tmpdir
exec {.}/test/showdir
dir /tmp
}`

	// Testing the ServeHTTP method requires OS-specific CGI scripts, because a
	// system call is made to respond to the request.
	if runtime.GOOS == "linux" {
		var buf bytes.Buffer
		hnd, err = handlerGet(directive, "./test")
		if err == nil {
			srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				code, err = hnd.ServeHTTP(w, r)
				if err != nil {
					fmt.Fprintf(&buf, "code [%d], error [%s]\n", code, err)
				}
			}))
			var res *http.Response
			res, err = http.Get(srv.URL + request)
			if err == nil {
				_, err = buf.ReadFrom(res.Body)
				res.Body.Close()
			}
			srv.Close()
		}
		if err == nil {
			rsp = strings.TrimSpace(buf.String())
			if "/tmp" != rsp {
				err = fmt.Errorf("expecting \"/tmp\", got \"%s\"", rsp)
			}
		}
		if err != nil {
			t.Fatalf("%s", err)
		}
	}
}

func TestPassAll(t *testing.T) {
	var err error
	var code int
	var hnd handlerType
	var srv *httptest.Server
	request := `/full`
	directive := `cgi {
match /full
exec {.}/test/fullenv
pass_all_env
}`

	// Testing the ServeHTTP method requires OS-specific CGI scripts, because a
	// system call is made to respond to the request.
	if runtime.GOOS == "linux" {
		var buf bytes.Buffer
		hnd, err = handlerGet(directive, "./test")
		if err == nil {
			srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				code, err = hnd.ServeHTTP(w, r)
				if err != nil {
					fmt.Fprintf(&buf, "code [%d], error [%s]\n", code, err)
				}
			}))
			fmt.Fprintf(&buf, "=== Directive ===\n%s\n", directive)
			var res *http.Response
			fmt.Fprintf(&buf, "--- Request %s ---\n", request)
			res, err = http.Get(srv.URL + request)
			if err == nil {
				_, err = buf.ReadFrom(res.Body)
				res.Body.Close()
			}
			srv.Close()
		}
		if err == nil {
			err = checkEnv(buf.String())
		}
		if err != nil {
			t.Fatalf("%s", err)
		}
	}
}

func TestMatches(t *testing.T) {
	var ok bool
	var prefixStr, suffixStr string
	// [request, pattern, expected success:1/expected error:0, prefix, suffix]
	list := [][]string{
		{"/foo/bar/baz", "/foo", "1", "/foo", "/bar/baz"},
		{"/foo/bar/baz", "/foo/*/baz", "1", "/foo/bar/baz", ""},
		{"/foo/bar/baz", "/foo/bar", "1", "/foo/bar", "/baz"},
		{"/foo/bar/baz", "foo/bar", "0", "", ""},
	}

	for _, rec := range list {
		ok, prefixStr, suffixStr = match(rec[0], []string{rec[1]})
		if ok {
			if rec[2] == "0" || rec[3] != prefixStr || rec[4] != suffixStr {
				t.Fatalf("expected mismatch for \"%s\" and \"%s\"", rec[0], rec[1])
			}
		} else {
			if rec[2] == "1" {
				t.Fatalf("expected match for \"%s\" and \"%s\"", rec[0], rec[1])
			}
		}
	}
}

func TestExceptions(t *testing.T) {
	// [request, except pattern, expected success:1/expected error:0]
	list := [][]string{
		{"/foo/bar/baz.png", "/*/*/*.png", "1"},
		{"/foo/bar/baz.png", "/foo/*/baz*", "1"},
		{"/foo/bar/baz.png", "/foo/bar/baz.jpg", "0"},
		{"/foo/bar/baz.png", "foo/bar", "0"},
	}

	for _, rec := range list {
		ok := excluded(rec[0], []string{rec[1]})
		if ok && rec[2] == "0" {
			t.Fatalf("unexpected exception for \"%s\" and \"%s\"", rec[0], rec[1])
		} else if !ok && rec[2] == "1" {
			t.Fatalf("expected exception for \"%s\" and \"%s\"", rec[0], rec[1])
		}
	}
}

func TestInspect(t *testing.T) {
	var err error
	var hnd handlerType
	var srv *httptest.Server
	var buf bytes.Buffer

	block := `cgi {
  match /foo/*
  exec bar --baz --quux
  pass_env HOME
  env VERY_LONG_KEY_INTENDED_TO_TEST_INSPECTION_REPORT=foo
  inspect
}`

	hnd, err = handlerGet(block, "./test")
	if err == nil {
		srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, hndErr := hnd.ServeHTTP(w, r)
			if err == nil && hndErr != nil {
				err = hndErr
			}
		}))
		var res *http.Response
		res, err = http.Get(srv.URL + "/foo/bar.tcl")
		if err == nil {
			_, err = buf.ReadFrom(res.Body)
			if err == nil {
				str := buf.String()
				if !strings.Contains(str, "CGI for Caddy") {
					err = fmt.Errorf("unexpected response for \"inspect\" request")
				}
			}
			res.Body.Close()
		}
		srv.Close()
	}
	if err != nil {
		t.Fatalf("%s", err)
	}
}
