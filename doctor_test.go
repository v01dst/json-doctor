package main

import (
	"strings"
	"testing"
)

const doc = `{"users":[{"name":"alice","age":30,"tags":["a","b"]},{"name":"bob","age":25}],"meta":{"count":2,"active":true}}`

func TestValidateOK(t *testing.T) {
	line, col, err := Validate([]byte(doc))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if line != 0 || col != 0 {
		t.Fatalf("want 0,0 got %d,%d", line, col)
	}
}

func TestValidateBadGivesPosition(t *testing.T) {
	bad := "{\n  \"a\": 1,\n  \"b\": tru\n}"
	line, col, err := Validate([]byte(bad))
	if err == nil {
		t.Fatal("want error")
	}
	if line != 4 {
		t.Fatalf("want line 4, got %d", line)
	}
	if col < 1 {
		t.Fatalf("want column >= 1, got %d", col)
	}
}

func TestValidateRejectsMultipleValues(t *testing.T) {
	if _, _, err := Validate([]byte(`{"a":1}{"b":2}`)); err == nil {
		t.Fatal("want error for multiple values")
	}
}

func TestPrettyAndMinify(t *testing.T) {
	prettyOut, err := Pretty([]byte(`{"a":1,"b":[1,2]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prettyOut, "\n  \"a\": 1") {
		t.Fatalf("pretty = %q", prettyOut)
	}

	minOut, err := Minify([]byte(prettyOut))
	if err != nil {
		t.Fatal(err)
	}
	if minOut != `{"a":1,"b":[1,2]}` {
		t.Fatalf("minify = %q", minOut)
	}
}

func TestAnalyze(t *testing.T) {
	res, err := Analyze([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if res.Keys != 9 {
		t.Fatalf("keys = %d", res.Keys)
	}
	if res.MaxDepth != 5 {
		t.Fatalf("maxDepth = %d", res.MaxDepth)
	}
	if res.Arrays != 2 {
		t.Fatalf("arrays = %d", res.Arrays)
	}
	if res.Objects != 4 {
		t.Fatalf("objects = %d", res.Objects)
	}
	if res.Types["string"] != 4 {
		t.Fatalf("strings = %d", res.Types["string"])
	}
	if len(res.TopKeys) != 2 {
		t.Fatalf("topKeys = %v", res.TopKeys)
	}
}

func TestQuery(t *testing.T) {
	val, err := Query([]byte(doc), "users.0.name")
	if err != nil {
		t.Fatal(err)
	}
	if val != "alice" {
		t.Fatalf("val = %v", val)
	}

	val, err = Query([]byte(doc), "users.0.tags.1")
	if err != nil {
		t.Fatal(err)
	}
	if val != "b" {
		t.Fatalf("val = %v", val)
	}

	val, err = Query([]byte(doc), "meta.active")
	if err != nil {
		t.Fatal(err)
	}
	if val != true {
		t.Fatalf("val = %v", val)
	}
}

func TestQueryErrors(t *testing.T) {
	if _, err := Query([]byte(doc), "users.5.name"); err == nil {
		t.Fatal("want out-of-range error")
	}
	if _, err := Query([]byte(doc), "meta.missing"); err == nil {
		t.Fatal("want missing key error")
	}
	if _, err := Query([]byte(doc), "meta.count.x"); err == nil {
		t.Fatal("want cannot-descend error")
	}
	if v, err := Query([]byte(doc), ""); err != nil || v == nil {
		t.Fatal("empty path should return whole doc")
	}
}

func TestAnalyzeNulls(t *testing.T) {
	res, err := Analyze([]byte(`{"a":null,"b":[null]}`))
	if err != nil {
		t.Fatal(err)
	}
	if res.Nulls != 2 {
		t.Fatalf("nulls = %d", res.Nulls)
	}
}

func TestSortKeys(t *testing.T) {
	unsorted := `{"z":1,"a":{"y":2,"b":3},"m":[ {"d":4,"c":5} ]}`
	sorted, err := SortKeys([]byte(unsorted))
	if err != nil {
		t.Fatal(err)
	}
	// a < m < z at top level
	aIdx := strings.Index(sorted, `"a"`)
	mIdx := strings.Index(sorted, `"m"`)
	zIdx := strings.Index(sorted, `"z"`)
	if !(aIdx < mIdx && mIdx < zIdx) {
		t.Fatalf("top-level keys not sorted: %s", sorted)
	}
	// nested b < y
	if strings.Index(sorted, `"b"`) > strings.Index(sorted, `"y"`) {
		t.Fatalf("nested keys not sorted: %s", sorted)
	}
	// array element keys c < d
	if strings.Index(sorted, `"c"`) > strings.Index(sorted, `"d"`) {
		t.Fatalf("array object keys not sorted: %s", sorted)
	}

	if _, err := SortKeys([]byte("not json")); err == nil {
		t.Fatal("invalid json should error")
	}
}

func TestMergeFiles(t *testing.T) {
	base := []byte(`{"a":1,"nested":{"x":1,"y":2},"keep":true}`)
	over := []byte(`{"nested":{"y":3,"z":4},"new":"here"}`)
	out, err := mergeFiles(base, over)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"a": 1`, `"x": 1`, `"y": 3`, `"z": 4`, `"new": "here"`, `"keep": true`} {
		if !strings.Contains(out, want) {
			t.Errorf("merged output missing %s:\n%s", want, out)
		}
	}
}

func TestMergeInvalidInput(t *testing.T) {
	if _, err := mergeFiles([]byte("nope"), []byte(`{"a":1}`)); err == nil {
		t.Fatal("invalid base should error")
	}
	if _, err := mergeFiles([]byte(`{"a":1}`), []byte("nope")); err == nil {
		t.Fatal("invalid overlay should error")
	}
}
