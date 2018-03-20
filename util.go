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
		for k, match := range r.matches {
			printf("  Match %d: %s\n", k, match)
		}
		for k, except := range r.exceptions {
			printf("  Except %d: %s\n", k, except)
		}
		printf("  Exe: %s\n", r.exe)
		for k, str := range r.args {
			printf("  Arg %d: %s\n", k, str)
		}
		for k, env := range r.envs {
			printf("  Env %d: %s=[%s]\n", k, env[0], env[1])
		}
		for k, str := range r.passEnvs {
			printf("  Pass env %d: %s\n", k, str)
		}
		for k, str := range r.emptyEnvs {
			printf("  Empty env %d: %s\n", k, str)
		}
	}
}
