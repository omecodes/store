package se

import (
	"bytes"
	"encoding/json"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/pb"
	"io"
	"strings"
)

func NewEngine(store Store) *Engine {
	return &Engine{
		store:     store,
		tokenizer: &textTokenizer{},
	}
}

type Engine struct {
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
	textAnalyzer := defaultTextAnalyzer()
	analyzedText := textAnalyzer(mapping.Text)

	stream := e.tokenizer.TokenizeText(analyzedText, mapping.PrefixMappingSize)

	var err error
	for {
		token := stream.Next()
		if token == "" {
			return nil
		}

		err = e.store.SaveWordMapping(token, mapping.ObjectId)
		if err != nil {
			return err
		}
	}
}

func (e *Engine) CreatePropertiesMapping(mapping *pb.PropertiesMapping) error {
	var props map[string]interface{}
	err := json.NewDecoder(bytes.NewBufferString(mapping.Json)).Decode(&props)
	if err != nil {
		return errors.Create(errors.BadRequest, err.Error())
	}

	for key, value := range props {
		if str, ok := value.(string); ok {
			textAnalyzer := propsMappingTextAnalyzer()
			props[key] = textAnalyzer(str)
		}
	}

	encoded, err := json.Marshal(props)
	if err != nil {
		return errors.Create(errors.Internal, err.Error(), errors.Info{Name: "engine", Details: "text not usable after analyze"})
	}

	return e.store.SavePropertiesMapping(mapping.ObjectId, string(encoded))
}

func (e *Engine) CreateNumberMapping(m *pb.NumberMapping) error {
	return e.store.SaveNumberMapping(m.Number, m.ObjectId)
}

func (e *Engine) DeleteObjectMappings(id string) error {
	return e.store.DeleteObjectMappings(id)
}

func (e *Engine) Search(query *pb.SearchQuery) ([]string, error) {
	c, err := e.store.Search(query)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cer := c.Close(); cer != nil {
			log.Error("cursor closing error", log.Err(cer))
		}
	}()

	sorter := &scoreRecords{}

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
