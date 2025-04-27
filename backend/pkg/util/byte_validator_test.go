package util

import (
	"strings"
	"testing"
)

// Test struct types for deserialization tests
type TestUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type TestProduct struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type TestNestedObject struct {
	Level1 struct {
		Level2 struct {
			Level3 struct {
				Level4 struct {
					Level5 struct {
						Level6 struct {
							Level7 struct {
								Level8 struct {
									Level9 struct {
										Level10 struct {
											Level11 struct {
												Value string `json:"value"`
											} `json:"level11"`
										} `json:"level10"`
									} `json:"level9"`
								} `json:"level8"`
							} `json:"level7"`
						} `json:"level6"`
					} `json:"level5"`
				} `json:"level4"`
			} `json:"level3"`
		} `json:"level2"`
	} `json:"level1"`
}

func TestNewByteValidator(t *testing.T) {
	validator := NewByteValidator(1024*1024, []string{}, []string{"$func", "$eval", "constructor", "prototype"}, 10)

	if validator.MaxSize != 1024*1024 {
		t.Errorf("Expected MaxSize to be %d, got %d", 1024*1024, validator.MaxSize)
	}

	if len(validator.AllowedTypes) != 0 {
		t.Errorf("Expected AllowedTypes to be empty, got %v", validator.AllowedTypes)
	}

	if len(validator.DisallowedKeys) != 4 {
		t.Errorf("Expected DisallowedKeys to have 4 items, got %d", len(validator.DisallowedKeys))
	}

	if validator.MaxDepth != 10 {
		t.Errorf("Expected MaxDepth to be %d, got %d", 10, validator.MaxDepth)
	}
}

func TestValidateBytes(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		config  func(*ByteValidator)
		wantErr bool
	}{
		{
			name:    "Nil data",
			data:    nil,
			config:  func(v *ByteValidator) {},
			wantErr: true,
		},
		{
			name:    "Empty data",
			data:    []byte{},
			config:  func(v *ByteValidator) {},
			wantErr: false,
		},
		{
			name:    "Valid data",
			data:    []byte(`{"name":"John","email":"john@example.com","age":30}`),
			config:  func(v *ByteValidator) {},
			wantErr: false,
		},
		{
			name:    "Data exceeding max size",
			data:    []byte(`{"name":"John","email":"john@example.com","age":30}`),
			config:  func(v *ByteValidator) { v.MaxSize = 10 },
			wantErr: true,
		},
		{
			name:    "Data with dangerous pattern",
			data:    []byte(`{"name":"<script>alert('XSS')</script>"}`),
			config:  func(v *ByteValidator) {},
			wantErr: true,
		},
		{
			name:    "Data with eval pattern",
			data:    []byte(`{"code":"eval(alert('Evil code'))"}`),
			config:  func(v *ByteValidator) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewByteValidator(1024*1024, []string{}, []string{"$func", "$eval", "constructor", "prototype"}, 10)
			tt.config(validator)

			err := validator.ValidateBytes(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateJSONSafety(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		config  func(*ByteValidator)
		wantErr bool
	}{
		{
			name:    "Valid JSON",
			json:    `{"name":"John","email":"john@example.com","age":30}`,
			config:  func(v *ByteValidator) {},
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			json:    `{"name":"John","email":}`,
			config:  func(v *ByteValidator) {},
			wantErr: true,
		},
		{
			name:    "JSON with disallowed key",
			json:    `{"constructor":"evil","name":"John"}`,
			config:  func(v *ByteValidator) {},
			wantErr: true,
		},
		{
			name:    "JSON with custom disallowed key",
			json:    `{"dangerous":"value"}`,
			config:  func(v *ByteValidator) { v.DisallowedKeys = []string{"dangerous"} },
			wantErr: true,
		},
		{
			name:    "JSON with dangerous key pattern",
			json:    `{"__proto__":"evil"}`,
			config:  func(v *ByteValidator) {},
			wantErr: true,
		},
		{
			name:    "JSON with excessive nesting",
			json:    `{"level1":{"level2":{"level3":{"level4":{"level5":{"level6":{"level7":{"level8":{"level9":{"level10":{"level11":{"value":"data"}}}}}}}}}}}`,
			config:  func(v *ByteValidator) { v.MaxDepth = 10 },
			wantErr: true,
		},
		{
			name:    "JSON with acceptable nesting",
			json:    `{"level1":{"level2":{"level3":{"level4":{"level5":{"value":"data"}}}}}}`,
			config:  func(v *ByteValidator) { v.MaxDepth = 10 },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewByteValidator(1024*1024, []string{}, []string{"$func", "$eval", "constructor", "prototype"}, 10)
			tt.config(validator)

			err := validator.validateJSONSafety([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("validateJSONSafety() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAndUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		target  interface{}
		config  func(*ByteValidator)
		wantErr bool
	}{
		{
			name:    "Valid user deserialization",
			json:    `{"name":"John","email":"john@example.com","age":30}`,
			target:  &TestUser{},
			config:  func(v *ByteValidator) {},
			wantErr: false,
		},
		{
			name:    "Valid product deserialization",
			json:    `{"id":1,"name":"Laptop","price":999.99}`,
			target:  &TestProduct{},
			config:  func(v *ByteValidator) {},
			wantErr: false,
		},
		{
			name:    "Type whitelist restriction - allowed",
			json:    `{"name":"John","email":"john@example.com","age":30}`,
			target:  &TestUser{},
			config:  func(v *ByteValidator) { v.AllowedTypes = []string{"*util.TestUser"} },
			wantErr: false,
		},
		{
			name:    "Type whitelist restriction - disallowed",
			json:    `{"id":1,"name":"Laptop","price":999.99}`,
			target:  &TestProduct{},
			config:  func(v *ByteValidator) { v.AllowedTypes = []string{"*util.TestUser"} },
			wantErr: true,
		},
		{
			name:    "Invalid field validation",
			json:    `{"name":"John","email":"john@example.com","age":30,"extraField":"value"}`,
			target:  &TestUser{},
			config:  func(v *ByteValidator) {},
			wantErr: true, // Should fail due to DisallowUnknownFields
		},
		{
			name:    "Dangerous JSON content",
			json:    `{"name":"<script>alert('XSS')</script>"}`,
			target:  &TestUser{},
			config:  func(v *ByteValidator) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewByteValidator(1024*1024, []string{}, []string{"$func", "$eval", "constructor", "prototype"}, 10)
			tt.config(validator)

			err := validator.ValidateAndUnmarshalJSON([]byte(tt.json), tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndUnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify deserialization was successful if expected
			if !tt.wantErr {
				// Check if target was populated correctly
				switch v := tt.target.(type) {
				case *TestUser:
					if tt.json == `{"name":"John","email":"john@example.com","age":30}` &&
						(v.Name != "John" || v.Email != "john@example.com" || v.Age != 30) {
						t.Errorf("Deserialization incorrect, got %+v", v)
					}
				case *TestProduct:
					if tt.json == `{"id":1,"name":"Laptop","price":999.99}` &&
						(v.ID != 1 || v.Name != "Laptop" || v.Price != 999.99) {
						t.Errorf("Deserialization incorrect, got %+v", v)
					}
				}
			}
		})
	}
}

func TestDeepNestedObjectDeserialization(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		config  func(*ByteValidator)
		wantErr bool
	}{
		{
			name:    "Nested object within max depth",
			json:    `{"level1":{"level2":{"level3":{"level4":{"level5":{"level6":{"level7":{"level8":{"level9":{"value":"data"}}}}}}}}}}`,
			config:  func(v *ByteValidator) { v.MaxDepth = 10 },
			wantErr: false,
		},
		{
			name:    "Nested object exceeding max depth",
			json:    `{"level1":{"level2":{"level3":{"level4":{"level5":{"level6":{"level7":{"level8":{"level9":{"level10":{"value":"data"}}}}}}}}}}`,
			config:  func(v *ByteValidator) { v.MaxDepth = 10 },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewByteValidator(1024*1024, []string{}, []string{"$func", "$eval", "constructor", "prototype"}, 10)
			tt.config(validator)

			var obj interface{}
			err := validator.ValidateAndUnmarshalJSON([]byte(tt.json), &obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("Deep nested object test error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCustomKeyPatterns(t *testing.T) {
	tests := []struct {
		name          string
		keyToTest     string
		wantDangerous bool
	}{
		{
			name:          "Normal key",
			keyToTest:     "name",
			wantDangerous: false,
		},
		{
			name:          "Double underscore key",
			keyToTest:     "__proto__",
			wantDangerous: true,
		},
		{
			name:          "Dollar sign key",
			keyToTest:     "$eval",
			wantDangerous: true,
		},
		{
			name:          "Underscore key",
			keyToTest:     "_hidden",
			wantDangerous: true,
		},
		{
			name:          "Special character key",
			keyToTest:     "normal-key",
			wantDangerous: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDangerousKey(tt.keyToTest)
			if result != tt.wantDangerous {
				t.Errorf("isDangerousKey(%s) = %v, want %v", tt.keyToTest, result, tt.wantDangerous)
			}
		})
	}
}

func TestDangerousPatterns(t *testing.T) {
	tests := []struct {
		name          string
		data          []byte
		wantDangerous bool
	}{
		{
			name:          "Safe content",
			data:          []byte(`{"name":"John","email":"john@example.com"}`),
			wantDangerous: false,
		},
		{
			name:          "Script tag",
			data:          []byte(`{"content":"<script>alert('XSS')</script>"}`),
			wantDangerous: true,
		},
		{
			name:          "Javascript URL",
			data:          []byte(`{"url":"javascript:alert(1)"}`),
			wantDangerous: true,
		},
		{
			name:          "Eval function",
			data:          []byte(`{"code":"eval('alert(1)')"}`),
			wantDangerous: true,
		},
		{
			name:          "Function constructor",
			data:          []byte(`{"code":"new Function('return alert(1)')"}`),
			wantDangerous: true,
		},
		{
			name:          "Proto pollution",
			data:          []byte(`{"__proto__":{"isAdmin":true}}`),
			wantDangerous: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsDangerousPatterns(tt.data)
			if result != tt.wantDangerous {
				t.Errorf("containsDangerousPatterns() = %v, want %v", result, tt.wantDangerous)
			}
		})
	}
}

// Integration test using multiple validator features together
func TestIntegrationScenarios(t *testing.T) {
	// Create a validator with custom configuration
	validator := NewByteValidator(1024*1024, []string{}, []string{"$func", "$eval", "constructor", "prototype"}, 10)
	validator.MaxSize = 512                             // Small size for testing
	validator.AllowedTypes = []string{"*util.TestUser"} // Make sure to use the correct package name
	validator.DisallowedKeys = append(validator.DisallowedKeys, "password", "secretKey")
	validator.MaxDepth = 3

	// Test 1: Valid data within constraints
	validJSON := `{"name":"Alice","email":"alice@example.com","age":25}`
	var user TestUser
	err := validator.ValidateAndUnmarshalJSON([]byte(validJSON), &user)
	if err != nil {
		t.Errorf("Integration test 1 failed: %v", err)
	}

	// Test 2: Data too large
	largeJSON := generateLargeJSON(1000)
	err = validator.ValidateAndUnmarshalJSON([]byte(largeJSON), &user)
	if err == nil {
		t.Errorf("Integration test 2 failed: expected error for large data")
	}

	// Test 3: Disallowed type
	productJSON := `{"id":1,"name":"Laptop","price":999.99}`
	var product TestProduct
	err = validator.ValidateAndUnmarshalJSON([]byte(productJSON), &product)
	if err == nil {
		t.Errorf("Integration test 3 failed: expected error for disallowed type")
	}

	// Test 4: Disallowed key
	passwordJSON := `{"name":"Bob","email":"bob@example.com","password":"secret123"}`
	err = validator.ValidateAndUnmarshalJSON([]byte(passwordJSON), &user)
	if err == nil {
		t.Errorf("Integration test 4 failed: expected error for disallowed key")
	}

	// Test 5: Excessive nesting
	nestedJSON := `{"level1":{"level2":{"level3":{"level4":{"value":"too deep"}}}}}`
	err = validator.ValidateAndUnmarshalJSON([]byte(nestedJSON), &map[string]interface{}{})
	if err == nil {
		t.Errorf("Integration test 5 failed: expected error for excessive nesting")
	}
}

// Helper function to generate a large JSON string
func generateLargeJSON(size int) string {
	padding := strings.Repeat("x", size)
	return `{"name":"` + padding + `"}`
}
