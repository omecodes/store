package cenv

import (
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	"strings"
	"time"
)

func stringsMatchOverload() *functions.Overload {
	return &functions.Overload{
		Operator: "string_match_string",
		Binary: func(lhs ref.Val, rhs ref.Val) ref.Val {
			pattern := lhs.Value().(string)
			candidate := rhs.Value().(string)

			patternParts := strings.Split(pattern, " ")
			candidateParts := strings.Split(candidate, " ")

			for _, patternItem := range patternParts {
				foundMatch := false
				for _, candidateItem := range candidateParts {
					patternItem = strings.ToLower(patternItem)
					candidateItem = strings.ToLower(candidateItem)

					if patternItem == candidateItem || strings.Contains(candidateItem, patternItem) {
						foundMatch = true
					}
				}

				if !foundMatch {
					return types.DefaultTypeAdapter.NativeToValue(false)
				}
			}
			return types.DefaultTypeAdapter.NativeToValue(true)
		},
	}
}

func nowOverload() *functions.Overload {
	return &functions.Overload{
		Operator: "now_uint",
		Function: func(values ...ref.Val) ref.Val {
			return types.DefaultTypeAdapter.NativeToValue(time.Now().UnixNano() / 1e6)
		},
	}
}
