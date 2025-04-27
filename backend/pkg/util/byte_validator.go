package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"unicode/utf8"
)

// ByteValidator provides a struct for byte array validation and safe deserialization
type ByteValidator struct {
	MaxSize        int      // Maximum allowed byte size
	AllowedTypes   []string // List of allowed type names for deserialization
	DisallowedKeys []string // List of disallowed keys (for JSON objects)
	MaxDepth       int      // Maximum allowed nesting depth
}

// NewByteValidator creates a new byte validator instance
func NewByteValidator(maxSize int, allowedTypes []string, disallowedKeys []string, maxDepth int) *ByteValidator {
	return &ByteValidator{
		MaxSize:        maxSize,        // Default 1MB
		AllowedTypes:   allowedTypes,   // Default empty list means no type restrictions
		DisallowedKeys: disallowedKeys, // Default dangerous keys to block
		MaxDepth:       maxDepth,       // Default maximum nesting depth
	}
}

// ValidateAndUnmarshalJSON validates byte array and safely deserializes it to JSON
func (v *ByteValidator) ValidateAndUnmarshalJSON(data []byte, target interface{}) error {
	// Basic validation
	if err := v.ValidateBytes(data); err != nil {
		return err
	}

	// Check JSON structure safety
	if err := v.validateJSONSafety(data); err != nil {
		return err
	}

	// Check if target object type is in the allowed list
	if len(v.AllowedTypes) > 0 {
		targetType := reflect.TypeOf(target).String()
		allowed := false
		for _, t := range v.AllowedTypes {
			if t == targetType {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("deserialization to type %s is not allowed", targetType)
		}
	}

	// Perform deserialization
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields() // Prevent unknown field injection

	return decoder.Decode(target)
}

// ValidateBytes performs basic byte validation
func (v *ByteValidator) ValidateBytes(data []byte) error {
	// Check for empty data
	if data == nil {
		return errors.New("input data is empty")
	}

	// Check size
	if len(data) > v.MaxSize {
		return fmt.Errorf("input data exceeds maximum allowed size of %d bytes", v.MaxSize)
	}

	// Check for invalid UTF-8 sequences
	if !utf8Valid(data) {
		return errors.New("input contains invalid UTF-8 sequences")
	}

	// Check for dangerous content
	if containsDangerousPatterns(data) {
		return errors.New("input contains potentially dangerous patterns")
	}

	return nil
}

// validateJSONSafety checks if JSON contains unsafe structures
func (v *ByteValidator) validateJSONSafety(data []byte) error {
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("invalid JSON format: %v", err)
	}

	// Check JSON structure depth and dangerous keys
	return v.checkJSONDepthAndKeys(jsonData, 0)
}

// checkJSONDepthAndKeys recursively checks JSON depth and dangerous keys
func (v *ByteValidator) checkJSONDepthAndKeys(data interface{}, depth int) error {
	// Check maximum depth
	if depth > v.MaxDepth {
		return fmt.Errorf("JSON nesting depth exceeds maximum allowed depth of %d", v.MaxDepth)
	}

	switch value := data.(type) {
	case map[string]interface{}:
		// Check for dangerous keys
		for key := range value {
			for _, disallowed := range v.DisallowedKeys {
				if key == disallowed {
					return fmt.Errorf("found disallowed key: %s", key)
				}
			}

			// Check for dangerous key patterns using regex
			if isDangerousKey(key) {
				return fmt.Errorf("found potentially dangerous key format: %s", key)
			}

			// Recursively check values
			if err := v.checkJSONDepthAndKeys(value[key], depth+1); err != nil {
				return err
			}
		}
	case []interface{}:
		// Recursively check array elements
		for _, item := range value {
			if err := v.checkJSONDepthAndKeys(item, depth+1); err != nil {
				return err
			}
		}
	}

	return nil
}

// utf8Valid checks if byte sequence is valid UTF-8
func utf8Valid(data []byte) bool {
	return utf8.Valid(data)
}

// containsDangerousPatterns checks for dangerous patterns
func containsDangerousPatterns(data []byte) bool {
	// Check for common dangerous patterns like script tags, expression injections, etc.
	dangerousPatterns := []string{
		`<script`,
		`javascript:`,
		`eval\(`,
		`Function\(`,
		`__proto__`,
		`__defineGetter__`,
		`__defineSetter__`,
	}

	for _, pattern := range dangerousPatterns {
		re := regexp.MustCompile("(?i)" + pattern)
		if re.Match(data) {
			return true
		}
	}
	return false
}

// isDangerousKey checks if key name has dangerous patterns
func isDangerousKey(key string) bool {
	dangerousKeyPatterns := []string{
		`^__.*__$`,    // Keys surrounded by double underscores
		`^(\$|_)\w+$`, // Special keys starting with $ or _
	}

	for _, pattern := range dangerousKeyPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(key) {
			return true
		}
	}
	return false
}
