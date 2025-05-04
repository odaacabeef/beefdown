package syllables

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// DatamuseWord represents the relevant fields from the Datamuse API response
type DatamuseWord struct {
	Word         string   `json:"word"`
	NumSyllables int      `json:"numSyllables"`
	Tags         []string `json:"tags"`
}

// SyllableCache provides thread-safe caching of syllable patterns
type SyllableCache struct {
	cache map[string][]string
	mu    sync.RWMutex
}

var globalCache = &SyllableCache{
	cache: make(map[string][]string),
}

// GetSyllables returns the syllables for a word, using cache or API
func GetSyllables(word string) []string {
	word = strings.ToLower(word)

	if word == "" {
		return []string{}
	}

	// Check cache first
	globalCache.mu.RLock()
	if syllables, ok := globalCache.cache[word]; ok {
		globalCache.mu.RUnlock()
		return syllables
	}
	globalCache.mu.RUnlock()

	// Query Datamuse API
	url := fmt.Sprintf("https://api.datamuse.com/words?sp=%s&md=s", word)
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("API error for word %q: %v", word, err)
		return []string{word} // Fallback to whole word
	}
	defer resp.Body.Close()

	var words []DatamuseWord
	if err := json.NewDecoder(resp.Body).Decode(&words); err != nil {
		log.Printf("JSON decode error for word %q: %v", word, err)
		return []string{word}
	}

	// Find exact match
	for _, w := range words {
		if w.Word == word {
			// Split word into syllables based on syllable count
			syllables := splitBySyllableCount(word, w.NumSyllables)

			// Cache the result
			globalCache.mu.Lock()
			globalCache.cache[word] = syllables
			globalCache.mu.Unlock()

			return syllables
		}
	}

	return []string{word}
}

// splitBySyllableCount splits a word into the specified number of syllables
func splitBySyllableCount(word string, numSyllables int) []string {
	if numSyllables <= 1 {
		return []string{word}
	}

	// Find vowel groups
	type vowelGroup struct {
		start int
		end   int
	}
	var groups []vowelGroup
	inVowel := false
	start := 0

	for i, r := range word {
		isVowel := strings.ContainsRune("aeiouy", r)
		if isVowel && !inVowel {
			start = i
			inVowel = true
		} else if !isVowel && inVowel {
			groups = append(groups, vowelGroup{start, i})
			inVowel = false
		}
	}
	if inVowel {
		groups = append(groups, vowelGroup{start, len(word)})
	}

	if len(groups) < numSyllables {
		return []string{word}
	}

	// Split into syllables
	syllables := make([]string, 0, numSyllables)
	charsPerSyllable := len(word) / numSyllables
	lastEnd := 0

	for i := 0; i < len(groups)-1; i++ {
		if len(syllables) == numSyllables-1 {
			break
		}

		// Find best split point
		nextGroup := groups[i+1]
		if nextGroup.start-lastEnd >= charsPerSyllable {
			splitPoint := groups[i].end
			syllables = append(syllables, word[lastEnd:splitPoint])
			lastEnd = splitPoint
		}
	}

	// Add remaining part as last syllable
	if lastEnd < len(word) {
		syllables = append(syllables, word[lastEnd:])
	}

	// If we couldn't split properly, return the whole word
	if len(syllables) != numSyllables {
		return []string{word}
	}

	return syllables
}
