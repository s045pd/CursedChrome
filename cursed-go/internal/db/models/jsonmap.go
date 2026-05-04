package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONMap is a generic JSON object backed by Postgres JSON column.
// Mirrors Sequelize DataTypes.JSON used heavily by the bots table.
type JSONMap map[string]any

// Value implements driver.Valuer.
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

// Scan implements sql.Scanner.
func (m *JSONMap) Scan(src any) error {
	if src == nil {
		*m = JSONMap{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("JSONMap: unsupported scan source")
	}
	if len(b) == 0 {
		*m = JSONMap{}
		return nil
	}
	return json.Unmarshal(b, m)
}

// JSONArray is a generic JSON array.
type JSONArray []any

// Value implements driver.Valuer.
func (a JSONArray) Value() (driver.Value, error) {
	if a == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(a)
}

// Scan implements sql.Scanner.
func (a *JSONArray) Scan(src any) error {
	if src == nil {
		*a = JSONArray{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("JSONArray: unsupported scan source")
	}
	if len(b) == 0 {
		*a = JSONArray{}
		return nil
	}
	return json.Unmarshal(b, a)
}
