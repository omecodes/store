package cenv

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/interpreter/functions"
	"strings"
)

func GetProgram(env *cel.Env, expression string) (cel.Program, error) {
	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	var opts []cel.ProgramOption
	var overloads []*functions.Overload

	if strings.Contains(expression, "match(") {
		overloads = append(overloads, stringsMatchOverload())
	}

	if strings.Contains(expression, "now()") {
		overloads = append(overloads, nowOverload())
	}

	if len(overloads) > 0 {
		opts = append(opts, cel.Functions(overloads...))
	}

	return env.Program(ast, opts...)
}
