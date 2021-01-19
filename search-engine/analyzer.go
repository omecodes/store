package se

import (
	"fmt"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"strings"
	"unicode"
)

type Analyzer func(in string) (out string)

type ChainAnalyzer func(analyzer Analyzer) Analyzer

func removePunctuation(analyzer Analyzer) Analyzer {
	return func(in string) string {
		var result strings.Builder
		for i := 0; i < len(in); i++ {
			b := in[i]
			if ('a' <= b && b <= 'z') ||
				('A' <= b && b <= 'Z') ||
				b == ' ' {
				result.WriteByte(b)
			}
		}
		return analyzer(result.String())
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

func removeStopWords(analyzer Analyzer) Analyzer {
	return func(in string) string {
		var out = in
		for _, word := range stopWords {
			out = strings.Replace(out, fmt.Sprintf(" %s ", word), "", -1)
		}
		return analyzer(out)
	}
}

func defaultTextAnalyzer() Analyzer {
	a := strings.ToLower
	a = removeStopWords(a)
	a = removePunctuation(a)
	a = removeAccents(a)
	return a
}

func propsMappingTextAnalyzer() Analyzer {
	a := strings.ToLower
	a = removePunctuation(a)
	a = removeAccents(a)
	return a

}

func getQueryTextAnalyzer() Analyzer {
	a := strings.ToLower
	a = removePunctuation(a)
	a = removeAccents(a)
	return a
}
