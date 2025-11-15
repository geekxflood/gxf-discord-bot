package config

import (
	"reflect"
	"testing"
)

func Test_parseKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want []keyPart
	}{
		{
			name: "simple key",
			key:  "foo",
			want: []keyPart{{key: "foo", index: -1}},
		},
		{
			name: "nested key",
			key:  "foo.bar",
			want: []keyPart{
				{key: "foo", index: -1},
				{key: "bar", index: -1},
			},
		},
		{
			name: "array access",
			key:  "actions[0]",
			want: []keyPart{{key: "actions", index: 0}},
		},
		{
			name: "array with nested",
			key:  "actions[0].name",
			want: []keyPart{
				{key: "actions", index: 0},
				{key: "name", index: -1},
			},
		},
		{
			name: "multiple arrays",
			key:  "servers[1].endpoints[2].url",
			want: []keyPart{
				{key: "servers", index: 1},
				{key: "endpoints", index: 2},
				{key: "url", index: -1},
			},
		},
		{
			name: "deep nesting",
			key:  "a.b.c.d.e",
			want: []keyPart{
				{key: "a", index: -1},
				{key: "b", index: -1},
				{key: "c", index: -1},
				{key: "d", index: -1},
				{key: "e", index: -1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseKey(tt.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getNestedValue(t *testing.T) {
	data := map[string]interface{}{
		"simple": "value",
		"nested": map[string]interface{}{
			"key": "nested-value",
		},
		"array": []interface{}{
			"item0",
			"item1",
			"item2",
		},
		"objects": []interface{}{
			map[string]interface{}{
				"name": "first",
				"id":   1,
			},
			map[string]interface{}{
				"name": "second",
				"id":   2,
			},
		},
		"deep": map[string]interface{}{
			"nested": map[string]interface{}{
				"array": []interface{}{
					map[string]interface{}{
						"value": "found",
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		key      string
		wantVal  interface{}
		wantOk   bool
	}{
		{
			name:    "simple key",
			key:     "simple",
			wantVal: "value",
			wantOk:  true,
		},
		{
			name:    "nested key",
			key:     "nested.key",
			wantVal: "nested-value",
			wantOk:  true,
		},
		{
			name:    "array index",
			key:     "array[0]",
			wantVal: "item0",
			wantOk:  true,
		},
		{
			name:    "array index 2",
			key:     "array[2]",
			wantVal: "item2",
			wantOk:  true,
		},
		{
			name:    "object in array",
			key:     "objects[0].name",
			wantVal: "first",
			wantOk:  true,
		},
		{
			name:    "object in array with int",
			key:     "objects[1].id",
			wantVal: 2,
			wantOk:  true,
		},
		{
			name:    "deep nested array",
			key:     "deep.nested.array[0].value",
			wantVal: "found",
			wantOk:  true,
		},
		{
			name:    "missing key",
			key:     "missing",
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "missing nested",
			key:     "nested.missing",
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "out of bounds array",
			key:     "array[99]",
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "negative array index",
			key:     "array[-1]",
			wantVal: nil,
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := getNestedValue(data, tt.key)
			if gotOk != tt.wantOk {
				t.Errorf("getNestedValue() ok = %v, want %v", gotOk, tt.wantOk)
				return
			}
			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t.Errorf("getNestedValue() val = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}
