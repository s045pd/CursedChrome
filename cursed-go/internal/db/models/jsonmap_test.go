package models

import (
	"reflect"
	"testing"
)

func TestJSONMap_RoundTrip(t *testing.T) {
	m := JSONMap{"a": "b", "n": float64(1)}
	v, err := m.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out JSONMap
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m, out) {
		t.Errorf("round-trip mismatch: %v != %v", m, out)
	}
}

func TestJSONMap_NilDefaultsToEmpty(t *testing.T) {
	var m JSONMap
	v, err := m.Value()
	if err != nil {
		t.Fatal(err)
	}
	if string(v.([]byte)) != "{}" {
		t.Errorf("nil JSONMap.Value() = %s, want {}", v)
	}

	if err := (&m).Scan(nil); err != nil {
		t.Fatal(err)
	}
	if len(m) != 0 {
		t.Errorf("Scan(nil) produced non-empty map: %v", m)
	}
}

func TestJSONArray_RoundTrip(t *testing.T) {
	a := JSONArray{"x", float64(2), true}
	v, err := a.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out JSONArray
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, out) {
		t.Errorf("round-trip mismatch: %v != %v", a, out)
	}
}

func TestJSONArray_NilDefaults(t *testing.T) {
	var a JSONArray
	v, err := a.Value()
	if err != nil {
		t.Fatal(err)
	}
	if string(v.([]byte)) != "[]" {
		t.Errorf("nil JSONArray.Value() = %s, want []", v)
	}

	if err := (&a).Scan(nil); err != nil {
		t.Fatal(err)
	}
	if len(a) != 0 {
		t.Errorf("Scan(nil) produced non-empty array: %v", a)
	}
}

func TestJSONMap_ScanStringInput(t *testing.T) {
	var m JSONMap
	if err := m.Scan(`{"k":"v"}`); err != nil {
		t.Fatal(err)
	}
	if m["k"] != "v" {
		t.Errorf("expected k=v, got %v", m)
	}
}

func TestJSONMap_ScanInvalidType(t *testing.T) {
	var m JSONMap
	if err := m.Scan(42); err == nil {
		t.Error("expected error for non string/[]byte source")
	}
}
