package cgi

import (
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// handlerType is a middleware type that can handle CGI requests
type handlerType struct {
	next  httpserver.Handler
	rules []ruleType
	root  string // same as root, but absolute path
}

// appType represents an application with arguments to execute when a rule is
// matched
type appType struct {
	// Glob patterns to match in order to apply rule
	matches []string // glob patterns
	// Name of executable script or binary
	exe string
	// Arguments to submit to executable
	args []string
	// Environment key value pairs ([0]: key, [1]: value) for this particular app
	// envs []keyValType
	envs [][2]string
	// Environment keys to pass through for this particular app
	passEnvs []string
}

// ruleType represents a CGI handling rule; it is parsed from the cgi directive
// in the Caddyfile
type ruleType struct {
	// Applications
	apps []appType
	// Environment key value pairs ([0]: key, [1]: value) for this particular app
	envs [][2]string
	// Environment keys to pass through for all apps
	passEnvs []string
}
