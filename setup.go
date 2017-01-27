package cgi

import (
	"path/filepath"
	"strings"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("cgi", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

// configureServer processes the tokens collected from the Caddy configuration
// file for the "cgi" directives and, if successful, inserts the cgi handler
// into the middleware chain.
func configureServer(ctrl *caddy.Controller, cfg *httpserver.SiteConfig) (err error) {
	var root string

	root, err = filepath.Abs(cfg.Root)
	if err == nil {
		var rules []ruleType
		rules, err = cgiParse(ctrl)
		if err == nil {
			cfg.AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
				return handlerType{
					next:  next,
					rules: rules,
					root:  root,
				}
			})
		}
	}
	return
}

// setup configures a new CGI middleware instance with the specified filesystem
// root
func setup(c *caddy.Controller) (err error) {
	return configureServer(c, httpserver.GetConfig(c))
}

// parseApp parses the brace-block following an "app" directive
func parseApp(c *caddy.Controller) (app appType, err error) {
	if c.Next() {
		if c.Val() == "{" {
			loop := true
			for err == nil && loop && c.Next() {
				val := c.Val()
				args := c.RemainingArgs()
				switch val {
				case "exec":
					if len(args) > 0 {
						app.exe = args[0]
						app.args = append(app.args, args[1:]...)
					} else {
						errorf("expecting at least one argument to follow \"exec\"")
					}
				case "match":
					if len(args) > 0 {
						app.matches = append(app.matches, args...)
					} else {
						errorf("expecting at least one argument to follow \"match\"")
					}
				case "env":
					var list [][2]string
					list, err = parseEnv(c, args)
					if err == nil {
						app.envs = append(app.envs, list...)
					}
				case "pass_env":
					app.passEnvs = append(app.passEnvs, args...)
				case "}":
					loop = false
				}
			}
		} else {
			errorf("expecting \"{\" in app block, got \"%s\"", c.Val())
		}
	} else {
		err = errorf("expecting brace block to follow \"app\"")
	}
	if err == nil {
		if len(app.matches) == 0 {
			err = errorf("at least one pattern match must be specified for app")
		} else if app.exe == "" {
			err = errorf("an executable must be specified in app block")
		}
	}
	return
}

// parseEnv parses a list of "key = value" pairs on a line
func parseEnv(c *caddy.Controller, args []string) (list [][2]string, err error) {
	count := len(args)
	for j := 0; j < count && err == nil; j++ {
		pair := strings.SplitN(args[j], "=", 2)
		if len(pair) == 2 {
			list = append(list, [2]string{
				strings.TrimSpace(pair[0]),
				strings.TrimSpace(pair[1]),
			})
		} else {
			err = c.Errf("expecting key=value format, got \"%s\"", args[j])
		}
	}
	return
}

// parseBlock parses the advance brace-block form of a "cgi" configuration
// directive
func parseBlock(c *caddy.Controller) (rule ruleType, err error) {
	if c.Next() {
		if c.Val() == "{" {
			loop := true
			for err == nil && loop && c.Next() {
				val := c.Val()
				args := c.RemainingArgs()
				switch val {
				case "app":
					var app appType
					switch len(args) {
					case 0: // brace block follows
						app, err = parseApp(c)
						if err == nil {
							rule.apps = append(rule.apps, app)
						}
					case 1:
						err = errorf("expecting \"app\" to follow simple syntax or advanced brace block syntax")
					default:
						app.matches = []string{args[0]}
						app.exe = args[1]
						app.args = append(app.args, args[2:]...)
						rule.apps = append(rule.apps, app)
					}
				case "env":
					var list [][2]string
					list, err = parseEnv(c, args)
					if err == nil {
						rule.envs = append(rule.envs, list...)
					}
				case "pass_env":
					rule.passEnvs = append(rule.passEnvs, args...)
				case "}":
					loop = false
				}
			}
			if len(rule.apps) == 0 {
				err = errorf("block must contain at least one application")
			}
		} else {
			err = errorf("expecting \"{\", got \"%s\"", c.Val())
		}
	} else {
		err = errorf("expecting brace block directive")
	}
	return
}

// cgiParse parses one or more "cgi" configuration directives
func cgiParse(c *caddy.Controller) (rules []ruleType, err error) {
	for err == nil && c.Next() {
		val := c.Val()
		args := c.RemainingArgs()
		if val == "cgi" {
			if len(args) == 0 { // advanced brace-block syntax
				var rule ruleType
				rule, err = parseBlock(c)
				if err == nil {
					rules = append(rules, rule)
				}
			} else if len(args) >= 2 { // simple one-line syntax
				rules = append(rules, ruleType{apps: []appType{{
					matches: []string{args[0]},
					exe:     args[1],
					args:    args[2:],
				}}})
			} else {
				err = errorf("expecting at least 2 arguments for simple directive, got %d", len(args))
			}
		} else {
			err = errorf("expecting \"cgi\", got \"%s\"", val)
		}
	}
	return
}
