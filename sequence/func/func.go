package funcpkg

import (
	metaparser "github.com/odaacabeef/beefdown/sequence/parsers/metadata"
)

// Func interface for all generative functions
type Func interface {
	Generate() ([]string, error) // Generates and returns step strings
}

// Factory creates a Func from metadata and parameters
type Factory func(meta metaparser.PartMetadata, params map[string]interface{}) (Func, error)

// Global registry of func types
var registry = make(map[string]Factory)

// Register registers a func type with the given factory
func Register(name string, factory Factory) {
	registry[name] = factory
}

// Get retrieves a func factory by name
func Get(name string) (Factory, bool) {
	factory, ok := registry[name]
	return factory, ok
}

// Helper functions for extracting typed parameters from generic map

func getStringParam(params map[string]interface{}, key string) (string, bool) {
	if val, ok := params[key]; ok {
		if node, ok := val.(*metaparser.StringNode); ok {
			return node.Value, true
		}
	}
	return "", false
}

func getIntParam(params map[string]interface{}, key string) (int, bool) {
	if val, ok := params[key]; ok {
		if node, ok := val.(*metaparser.NumberNode); ok {
			return int(node.Value), true
		}
	}
	return 0, false
}

func getFloatParam(params map[string]interface{}, key string) (float64, bool) {
	if val, ok := params[key]; ok {
		if node, ok := val.(*metaparser.NumberNode); ok {
			return node.Value, true
		}
	}
	return 0, false
}
