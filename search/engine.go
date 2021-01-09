package search

import "github.com/omecodes/store/pb"

type Engine struct {
	store           Store
	chainedAnalyzer Analyzer
}

func (e *Engine) Search(expression *pb.BooleanExp) (Cursor, error) {
	return nil, nil
}
