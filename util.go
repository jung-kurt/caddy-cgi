package cgi

import (
	"fmt"
	"strings"
)

var (
	errorf  = fmt.Errorf
	join    = strings.Join
	printf  = fmt.Printf
	sprintf = fmt.Sprintf
	trim    = strings.TrimSpace
)

func unused(args ...interface{}) {
}

// printRules displays rules on standard output
func printRules(rules []ruleType) {
	for j, r := range rules {
		printf("Rule %d\n", j)
		for k, app := range r.apps {
			printf("  App %d\n", k)
			for l, match := range app.matches {
				printf("    Match %d: %s\n", l, match)
			}
			printf("    Exe: %s\n", app.exe)
			for l, str := range app.args {
				printf("    Arg %d: %s\n", l, str)
			}
			for l, env := range app.envs {
				printf("    Env %d: %s=[%s]\n", l, env[0], env[1])
			}
			for l, str := range app.passEnvs {
				printf("    Pass env %d: %s\n", l, str)
			}
		}
		for l, env := range r.envs {
			printf("  Env %d: %s=[%s]\n", l, env[0], env[1])
		}
		for l, str := range r.passEnvs {
			printf("  Pass env %d: %s\n", l, str)
		}
	}
}
