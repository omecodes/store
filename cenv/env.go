package cenv

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func ACLEnv() (*cel.Env, error) {
	return cel.NewEnv(
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
}

func SearchEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Declarations(
			decls.NewVar("o", decls.NewMapType(decls.String, decls.Dyn)),
			decls.NewFunction("match",
				decls.NewOverload(
					"string_match_string",
					[]*expr.Type{decls.String, decls.String}, decls.Bool,
				),
			),
			decls.NewFunction("now",
				decls.NewOverload(
					"now_uint",
					[]*expr.Type{}, decls.Uint,
				),
			),
		),
	)
}
