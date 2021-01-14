package se

type scoreRecords struct {
	scores map[string]int
}

func (s *scoreRecords) append(ids []string) {
	if s.scores == nil {
		s.scores = make(map[string]int)
	}

	for _, id := range ids {
		score, found := s.scores[id]
		if found {
			s.scores[id] = score + 1
		} else {
			s.scores[id] = 0
		}
	}
}

func (s *scoreRecords) sorted() []string {
	var sortedIds []string
	var sortedValues []int

	for key, val := range s.scores {
		ind := 0
		for _, sk := range sortedValues {
			if val > sk {
				break
			}
			ind++
		}

		if ind == len(sortedValues) {
			sortedValues = append(sortedValues, val)
			sortedIds = append(sortedIds, key)

		} else if ind == 0 {
			sortedValues = append([]int{val}, sortedValues...)
			sortedIds = append([]string{key}, sortedIds...)

		} else {
			sortedValues = append(append(sortedValues[:ind], val), sortedValues[ind+1:]...)
			sortedIds = append(append(sortedIds[:ind], key), sortedIds[ind+1:]...)
		}
	}
	return sortedIds
}

type tokenMatchScorer func(string, string, *scoreRecords)

func containsScorer(pattern string) tokenMatchScorer {
	return func(token string, id string, records *scoreRecords) {}
}

func startsWithScorer(pattern string) tokenMatchScorer {
	return func(token string, id string, records *scoreRecords) {}
}

func equalsScorer(pattern string) tokenMatchScorer {
	return func(token string, id string, records *scoreRecords) {}
}

func endsWithScorer(pattern string) tokenMatchScorer {
	return func(token string, id string, records *scoreRecords) {}
}
