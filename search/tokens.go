package se

import "strings"

type TextTokenizer interface {
	TokenizeText(text string, originalMappedCount uint32) TokenStream
}

type textTokenizer struct {
	text string
}

func (t *textTokenizer) TokenizeText(text string, originalMappedCount uint32) TokenStream {
	tokens := strings.Split(text, " ")
	if len(text) < int(originalMappedCount) || originalMappedCount == 0 {
		tokens = append(tokens, text)
	} else {
		tokens = append(tokens, text[:originalMappedCount])
	}
	return &tokenStream{tokens: tokens}
}

type TokenStream interface {
	Flip()
	Next() string
}

type tokenStream struct {
	tokens []string
	cursor int
}

func (stream *tokenStream) Flip() {
	stream.cursor = 0
}

func (stream *tokenStream) Next() string {
	if stream.cursor < len(stream.tokens) {
		item := stream.tokens[stream.cursor]
		stream.cursor++
		return item
	}
	return ""
}
