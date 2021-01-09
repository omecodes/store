package se

import (
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"regexp"
	"strings"
	"unicode"
)

type Analyzer func(in string) (out string)

type ChainAnalyzer func(analyzer Analyzer) Analyzer

func removePunctuation(analyzer Analyzer) Analyzer {
	return func(in string) string {
		reg, err := regexp.Compile("[^a-zA-Z0-9]+")
		if err != nil {
			return in
		}
		out := reg.ReplaceAllString(in, "")
		return analyzer(out)
	}
}

func removeAccents(analyzer Analyzer) Analyzer {
	return func(in string) string {
		t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
		out, _, err := transform.String(t, in)
		if err != nil && out == "" {
			return in
		}
		return analyzer(out)
	}
}

func removeIgnored(analyzer Analyzer) Analyzer {
	return func(in string) string {
		var out = in
		for _, word := range stopWords {
			out = strings.Replace(out, word, "", -1)
		}
		return analyzer(out)
	}
}

func getTextAnalyzer() Analyzer {
	a := strings.ToLower
	a = removeIgnored(a)
	a = removePunctuation(a)
	a = removeAccents(a)
	return a
}
