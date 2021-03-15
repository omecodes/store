package cenv

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/interpreter/functions"
	"github.com/omecodes/libome/logs"
	_ "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"strings"
)

var aclEnv *cel.Env

func acl(opts ...cel.EnvOption) (*cel.Env, error) {
	if aclEnv == nil {
		var err error
		aclEnv, err = cel.NewEnv(
			opts...,
		)
		if err != nil {
			return nil, err
		}
	}
	return aclEnv, nil
}

func GetProgram(expression string, envOpts ...cel.EnvOption) (cel.Program, error) {
	env, err := acl(envOpts...)
	if err != nil {
		return nil, err
	}

	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		logs.Error("could not compile expression", logs.Details("issues", issues.String()), logs.Err(issues.Err()))
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
