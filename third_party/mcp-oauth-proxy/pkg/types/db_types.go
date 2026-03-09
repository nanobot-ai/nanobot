package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringSlice is a custom type for handling string slices in GORM
type StringSlice []string

// JSON is a custom type for handling JSON data in GORM
type JSON map[string]any

// Value implements the driver.Valuer interface for StringSlice
func (s StringSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// Scan implements the sql.Scanner interface for StringSlice
func (s *StringSlice) Scan(value any) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("cannot scan %T into StringSlice", value)
	}

	if len(data) == 0 {
		*s = []string{}
		return nil
	}

	return json.Unmarshal(data, s)
}

// Value implements the driver.Valuer interface for JSON
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}
	data, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// Scan implements the sql.Scanner interface for JSON
func (j *JSON) Scan(value any) error {
	if value == nil {
		*j = make(map[string]any)
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("cannot scan %T into JSON", value)
	}

	if len(data) == 0 {
		*j = make(map[string]any)
		return nil
	}

	return json.Unmarshal(data, j)
}
