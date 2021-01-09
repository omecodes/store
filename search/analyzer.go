package search

type Analyzer func(string) string

type ChainAnalyzer func(analyzer Analyzer) Analyzer
