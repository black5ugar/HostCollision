package similarity

import (
	"strings"
)

// Score computes a similarity score between two byte slices.
// The score is an integer value between 0 and 100.
func Score(a, b []byte) int {
	s1 := normalize(string(a))
	s2 := normalize(string(b))

	if len(s1) == 0 && len(s2) == 0 {
		return 100
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0
	}

	m1 := shingles(s1, 3)
	m2 := shingles(s2, 3)

	if len(m1) == 0 && len(m2) == 0 {
		return 100
	}

	var intersection, union int

	for k, v1 := range m1 {
		if v2, ok := m2[k]; ok {
			if v1 < v2 {
				intersection += v1
			} else {
				intersection += v2
			}
			if v1 > v2 {
				union += v1
			} else {
				union += v2
			}
		} else {
			union += v1
		}
	}

	for k, v2 := range m2 {
		if _, ok := m1[k]; !ok {
			union += v2
		}
	}

	if union == 0 {
		return 0
	}

	ratio := float64(intersection) / float64(union)
	score := int(ratio*100 + 0.5)
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// normalize prepares the string for similarity comparison.
func normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSpace(s)
	return s
}

// shingles builds a multiset of fixed-size substrings from the input string.
func shingles(s string, size int) map[string]int {
	result := make(map[string]int)
	if size <= 0 {
		return result
	}

	if len(s) <= size {
		result[s]++
		return result
	}

	for i := 0; i <= len(s)-size; i++ {
		token := s[i : i+size]
		result[token]++
	}
	return result
}
