package syllables

import (
	"testing"
	"time"
)

func TestGetSyllables(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		expected []string
	}{
		{
			name:     "simple word",
			word:     "hello",
			expected: []string{"he", "llo"},
		},
		{
			name:     "multi-syllable word",
			word:     "beautiful",
			expected: []string{"beau", "ti", "ful"},
		},
		{
			name:     "single syllable word",
			word:     "cat",
			expected: []string{"cat"},
		},
		{
			name:     "empty string",
			word:     "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSyllables(tt.word)
			if len(got) != len(tt.expected) {
				t.Errorf("GetSyllables(%q) returned %d syllables, want %d", tt.word, len(got), len(tt.expected))
				return
			}
			for i, syllable := range got {
				if syllable != tt.expected[i] {
					t.Errorf("GetSyllables(%q)[%d] = %q, want %q", tt.word, i, syllable, tt.expected[i])
				}
			}
		})
	}
}

func TestCaching(t *testing.T) {
	// Clear the cache
	globalCache.mu.Lock()
	globalCache.cache = make(map[string][]string)
	globalCache.mu.Unlock()

	word := "hello"

	// First call should hit the API
	start := time.Now()
	firstResult := GetSyllables(word)
	firstDuration := time.Since(start)

	// Second call should use cache
	start = time.Now()
	secondResult := GetSyllables(word)
	secondDuration := time.Since(start)

	// Verify results are the same
	if len(firstResult) != len(secondResult) {
		t.Errorf("Cached result differs from first result")
		return
	}
	for i, syllable := range firstResult {
		if syllable != secondResult[i] {
			t.Errorf("Cached result differs from first result at index %d", i)
		}
	}

	// Verify second call was faster (cached)
	if secondDuration >= firstDuration {
		t.Errorf("Cached call (%v) was not faster than API call (%v)", secondDuration, firstDuration)
	}
}

func TestFallback(t *testing.T) {
	// Test with a non-existent word
	word := "supercalifragilisticexpialidocious"
	result := GetSyllables(word)

	if len(result) != 1 || result[0] != word {
		t.Errorf("GetSyllables(%q) = %v, want [%q]", word, result, word)
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Test concurrent access to the cache
	words := []string{"hello", "beautiful", "computer", "rhythm"}
	done := make(chan bool)

	for _, word := range words {
		go func(w string) {
			GetSyllables(w)
			done <- true
		}(word)
	}

	// Wait for all goroutines to complete
	for range words {
		<-done
	}
}
