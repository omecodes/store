package se

import "strings"

type TextTokenizer interface {
	TokenizeText(text string) TokenStream
}

type textTokenizer struct {
	text string
}

func (t *textTokenizer) TokenizeText(text string) TokenStream {
	tokens := strings.Split(text, " ")
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
		return stream.tokens[stream.cursor]
	}
	return ""
}
