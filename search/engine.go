package se

import (
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/store/pb"
	"io"
	"strings"
)

func NewEngine(store Store) *Engine {
	return &Engine{
		store:     store,
		analyzer:  getMappingTextAnalyzer(),
		tokenizer: &textTokenizer{},
	}
}

type Engine struct {
	analyzer  Analyzer
	tokenizer TextTokenizer
	store     Store
}

func (e *Engine) Feed(msg *pb.MessageFeed) error {
	switch m := msg.Message.(type) {
	case *pb.MessageFeed_TextMapping:
		return e.CreateTextMapping(m.TextMapping)

	case *pb.MessageFeed_NumMapping:
		return e.CreateNumberMapping(m.NumMapping)

	case *pb.MessageFeed_Delete:
	}

	return nil
}

func (e *Engine) CreateTextMapping(mapping *pb.TextMapping) error {
	analyzedText := e.analyzer(mapping.Text)
	stream := e.tokenizer.TokenizeText(analyzedText)

	var err error
	for {
		token := stream.Next()
		if token == "" {
			return nil
		}

		err = e.store.SaveWordMapping(token, mapping.FieldName, mapping.ObjectId)
		if err != nil {
			return err
		}
	}
}

func (e *Engine) CreateNumberMapping(m *pb.NumberMapping) error {
	return e.store.SaveNumberMapping(m.Number, m.FieldName, m.ObjectId)
}

func (e *Engine) DeleteObjectMappings(id string) error {
	return e.store.DeleteObjectMappings(id)
}

func (e *Engine) Search(expression *pb.BooleanExp) ([]string, error) {
	c, err := e.store.Search(expression)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cer := c.Close(); cer != nil {
			log.Error("cursor closing error", log.Err(cer))
		}
	}()

	sorter := &scoreSorter{}

	for {
		value, err := c.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		ids := strings.Split(value, " ")
		sorter.append(ids)
	}

	return sorter.sorted(), nil
}
