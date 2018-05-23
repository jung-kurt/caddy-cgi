package cgi

import (
	"fmt"
	"testing"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func configure(expectErr bool, str string) (rules []ruleType, err error) {
	var c *caddy.Controller
	var mids []httpserver.Middleware
	var cfg *httpserver.SiteConfig
	var srvHnd httpserver.Handler
	var hnd handlerType
	var ok bool

	// printf("--- Directive begin --\n%s\n--- Directive end ---\n", str)
	c = caddy.NewTestController("http", str)
	err = setup(c)
	if err == nil {
		if !expectErr {
			cfg = httpserver.GetConfig(c)
			mids = cfg.Middleware()
			midLen := len(mids)
			if midLen > 0 {
				for j := 0; j < midLen && err == nil; j++ {
					srvHnd = mids[j](httpserver.EmptyNext)
					hnd, ok = srvHnd.(handlerType)
					if ok {
						rules = append(rules, hnd.rules...)
					} else {
						err = fmt.Errorf("expected middleware handler to be CGI handler")
					}
				}
			} else {
				err = fmt.Errorf("no middlewares present")
			}
		} else {
			err = fmt.Errorf("expected error but got succcess")
		}
	} else if expectErr {
		err = nil
	}
	return
}

// This examples demonstrates printing a CGI rule
func Example_rule() {
	var err error
	var rules []ruleType
	var strList = []string{
		`cgi {
  match *.lua *.luac
  except init.lua init.luac
  except utility.lua
  exec /usr/bin/lua /usr/local/cgi-bin/{match}
  pass_env LUA_PATH LUA_CPATH
  empty_env REMOTE_USER CONTENT_TYPE
 }`,
		`cgi {
  match *.py *.pyc
  exec /usr/bin/python -s
  env PYTHONSTARTUP=/usr/share/init.py
 }`,
		`cgi /fossil /var/www/fossil`,
		`cgi {
  match /report/week
  exec /var/www/report --mode=week
  env NO_BANANAS=YES "NAME = Don Quixote" 
  env MODE=DEV
  pass_env JWT_SECRET
}`,
	}
	for _, str := range strList {
		rules, err = configure(false, str)
		if err == nil {
			printRules(rules)
		} else {
			printf("%s\n", err)
		}
	}
	// Output:
	// Rule 0
	//   Match 0: *.lua
	//   Match 1: *.luac
	//   Except 0: init.lua
	//   Except 1: init.luac
	//   Except 2: utility.lua
	//   Exe: /usr/bin/lua
	//   Arg 0: /usr/local/cgi-bin/{match}
	//   Pass env 0: LUA_PATH
	//   Pass env 1: LUA_CPATH
	//   Empty env 0: REMOTE_USER
	//   Empty env 1: CONTENT_TYPE
	// Rule 0
	//   Match 0: *.py
	//   Match 1: *.pyc
	//   Exe: /usr/bin/python
	//   Arg 0: -s
	//   Env 0: PYTHONSTARTUP=[/usr/share/init.py]
	// Rule 0
	//   Match 0: /fossil
	//   Exe: /var/www/fossil
	// Rule 0
	//   Match 0: /report/week
	//   Exe: /var/www/report
	//   Arg 0: --mode=week
	//   Env 0: NO_BANANAS=[YES]
	//   Env 1: NAME=[Don Quixote]
	//   Env 2: MODE=[DEV]
	//   Pass env 0: JWT_SECRET
}

func TestSetup(t *testing.T) {
	var err error
	// Each of the following directives is submitted for parsing. If the string is
	// prefixed with "0:", it is expected to parse successfully. If it is prefixed
	// with "1:", an error is expected. The prefix is removed before parsing.
	var directiveList = []string{
		`0:cgi /report/daily /usr/bin/perl /usr/share/perl/report --mode=daily`,

		`0:cgi {
  match *.lua *.luac
  except init.lua
  exec /usr/bin/lua /usr/local/cgi-bin/{match}
  pass_env LUA_PATH LUA_CPATH
 }
 cgi {
   match *.py *.pyc
   exec /usr/bin/python -s /usr/local/cgi-bin/{match}
   env PYTHONSTARTUP=/usr/share/init.py
}
cgi {
  match /fossil 
  exec /var/www/fossil /usr/local/cgi-bin/fossil
}
cgi {
  match /report/week 
  exec /usr/local/cgi-bin/report --mode=week
  env NO_BANANAS=YES "NAME = Don Quixote" 
  env MODE=DEV
  pass_env JWT_SECRET
}`,

		`1:cgi {
  match /foo /foo/script -a
}`,

		`1:cgi {
  match /foo
}`,

		`1:cgi {
  match *.lua *.luac
}`,

		`1:cgi {
  match /*.pl
  exec
}`,

		`1:cgi {
  match /*.pl
  except
  exec /usr/bin/perl
}`,

		`0:cgi {
  match /*.pl
  except init.pl
  exec /usr/bin/perl
  inspect
}`,

		`1:cgi {
  match /*.pl
  except init.pl
  exec /usr/bin/perl
  inspect foo
}`,

		`0:cgi {
  match /*.pl
  except init.pl
  exec /usr/bin/perl
}`,

		`1:cgi {
  match
  exec /usr/bin/perl
}`,

		`1:cgi {
  match
`,

		`1:cgi {
  exec
`,

		`1:cgi /report/daily /usr/bin/perl /usr/share/perl/report --mode=daily
  xcgi /foo
  cgi /bar /usr/bin/bar -a -b -c`,

		`1:cgi`,

		`1:cgi
  env`,

		`1:cgi {
  env NO_BANANAS:MAYBE
 }`,

		`1:cgi {
  match
  match /foo /bin/foo`,

		`1:cgi {
  match /report
  exe /usr/local/bin/foo
  exe /usr/local/bin/bar
}`,

		`1:cgi /report/daily`,
	}

	for j := 0; j < len(directiveList) && err == nil; j++ {
		str := directiveList[j]
		_, err = configure(str[0:1] != "0", str[2:])
		if err != nil {
			printf("error [%s], str [%s]\n", err.Error(), str)
		}
	}
	if err != nil {
		t.Fatal(err)
	}
}
