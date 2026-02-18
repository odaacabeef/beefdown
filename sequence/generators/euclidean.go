package generators

import (
	"fmt"
	"math/rand"
	"strings"

	metaparser "github.com/odaacabeef/beefdown/sequence/parsers/metadata"
)

// Euclidean generates Euclidean rhythms using the Bjorklund algorithm
// Distributes pulses as evenly as possible across steps
type Euclidean struct {
	Pulses   int
	Steps    int
	Note     string // Single note or comma-separated pool
	Rotation int
	Seed     int64
}

func (e *Euclidean) Generate() ([]string, error) {
	if e.Pulses < 0 || e.Steps < 0 {
		return nil, fmt.Errorf("euclidean: pulses and steps must be non-negative")
	}
	if e.Pulses > e.Steps {
		return nil, fmt.Errorf("euclidean: pulses (%d) cannot exceed steps (%d)", e.Pulses, e.Steps)
	}
	if e.Note == "" {
		return nil, fmt.Errorf("euclidean: note is required")
	}

	// Generate Euclidean rhythm pattern
	pattern := bjorklund(e.Pulses, e.Steps)

	// Apply rotation if specified
	if e.Rotation != 0 {
		pattern = rotate(pattern, e.Rotation)
	}

	// Check if Note is a pool (contains commas)
	notePool := strings.Split(e.Note, ",")
	for i := range notePool {
		notePool[i] = strings.TrimSpace(notePool[i])
	}
	isPool := len(notePool) > 1

	// Initialize RNG if we have a pool
	var rng *rand.Rand
	if isPool {
		rng = rand.New(rand.NewSource(e.Seed))
	}

	// Convert pattern to steps
	var steps []string
	for _, pulse := range pattern {
		if pulse {
			var note string
			if isPool {
				// Randomly select from pool
				note = notePool[rng.Intn(len(notePool))]
			} else {
				// Use single note
				note = e.Note
			}
			steps = append(steps, fmt.Sprintf("%s:1", note))
		} else {
			steps = append(steps, "") // rest
		}
	}

	return steps, nil
}

// bjorklund implements the Bjorklund algorithm for generating Euclidean rhythms
// Returns a slice of bools where true = pulse, false = rest
// This uses a simple iterative approach based on the Euclidean algorithm
func bjorklund(pulses, steps int) []bool {
	if pulses == 0 || steps == 0 {
		result := make([]bool, steps)
		return result
	}

	if pulses >= steps {
		result := make([]bool, steps)
		for i := range result {
			result[i] = true
		}
		return result
	}

	// Use the Bresenham line algorithm approach for simplicity and clarity
	// This distributes pulses as evenly as possible
	pattern := make([]bool, steps)
	bucket := 0

	for i := 0; i < steps; i++ {
		bucket += pulses
		if bucket >= steps {
			bucket -= steps
			pattern[i] = true
		} else {
			pattern[i] = false
		}
	}

	return pattern
}

// rotate rotates a pattern by the specified amount
// Positive values rotate right, negative values rotate left
func rotate(pattern []bool, amount int) []bool {
	if len(pattern) == 0 {
		return pattern
	}

	// Normalize rotation amount to be within pattern length
	amount = amount % len(pattern)
	if amount < 0 {
		amount += len(pattern)
	}

	rotated := make([]bool, len(pattern))
	for i := range pattern {
		rotated[(i+amount)%len(pattern)] = pattern[i]
	}
	return rotated
}

func newEuclidean(meta metaparser.PartMetadata, params map[string]interface{}) (Generator, error) {
	pulses, ok := getIntParam(params, "pulses")
	if !ok {
		return nil, fmt.Errorf("euclidean: missing required parameter 'pulses'")
	}

	steps, ok := getIntParam(params, "steps")
	if !ok {
		return nil, fmt.Errorf("euclidean: missing required parameter 'steps'")
	}

	note, ok := getStringParam(params, "note")
	if !ok {
		// Try "notes" as alternative parameter name
		note, ok = getStringParam(params, "notes")
		if !ok {
			return nil, fmt.Errorf("euclidean: missing required parameter 'note' or 'notes'")
		}
	}

	rotation, _ := getIntParam(params, "rotation") // optional, defaults to 0
	seed, _ := getIntParam(params, "seed")         // optional, defaults to 0

	return &Euclidean{
		Pulses:   pulses,
		Steps:    steps,
		Note:     note,
		Rotation: rotation,
		Seed:     int64(seed),
	}, nil
}

func init() {
	Register("euclidean", newEuclidean)
}
