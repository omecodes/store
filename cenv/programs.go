package cenv

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/interpreter/functions"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"strings"
)

var aclEnv *cel.Env

func acl() (*cel.Env, error) {
	if aclEnv == nil {
		var err error
		aclEnv, err = cel.NewEnv(
			cel.Declarations(
				decls.NewVar("auth", decls.NewMapType(decls.String, decls.Dyn)),
				decls.NewVar("data", decls.NewMapType(decls.String, decls.Dyn)),
				decls.NewFunction("now",
					decls.NewOverload(
						"now_uint",
						[]*expr.Type{}, decls.Uint,
					),
				),
			),
		)
		if err != nil {
			return nil, err
		}
	}
	return aclEnv, nil
}

func GetProgram(expression string) (cel.Program, error) {
	env, err := acl()
	if err != nil {
		return nil, err
	}

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
