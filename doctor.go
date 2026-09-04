package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// AnalyzeResult holds structural statistics about a JSON document.
type AnalyzeResult struct {
	Keys     int            `json:"keys"`
	MaxDepth int            `json:"maxDepth"`
	Arrays   int            `json:"arrays"`
	Objects  int            `json:"objects"`
	Scalars  int            `json:"scalars"`
	Nulls    int            `json:"nulls"`
	Types    map[string]int `json:"types"`
	TopKeys  []string       `json:"topKeys"`
	ByteSize int            `json:"byteSize"`
}

// Validate checks JSON and returns the line/column of a syntax error, if any.
func Validate(input []byte) (line, col int, err error) {
	var v any
	if json.Valid(input) {
		// Fully decode to catch trailing garbage json.Valid may accept? No,
		// json.Valid is strict. Decode for type checking only.
		dec := json.NewDecoder(bytes.NewReader(input))
		dec.UseNumber()
		if dec.Decode(&v) != nil {
			return 0, 0, fmt.Errorf("invalid json")
		}
		return 0, 0, nil
	}
	var syn *json.SyntaxError
	dec := json.NewDecoder(bytes.NewReader(input))
	decErr := dec.Decode(&v)
	if decErr != nil {
		if e, ok := decErr.(*json.SyntaxError); ok {
			syn = e
		} else {
			return 0, 0, decErr
		}
	} else if dec.Decode(&v) == nil {
		return 0, 0, fmt.Errorf("multiple json values in input")
	} else if json.Valid(bytes.TrimSpace(input)) {
		return 0, 0, nil
	}
	if syn != nil {
		l, c := offsetToLineCol(input, int(syn.Offset))
		return l, c, syn
	}
	return 0, 0, fmt.Errorf("invalid json")
}

func offsetToLineCol(input []byte, offset int) (int, int) {
	if offset > len(input) {
		offset = len(input)
	}
	line := 1
	col := 1
	for i := 0; i < offset; i++ {
		if input[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

// Pretty reformats JSON with 2-space indentation.
func Pretty(input []byte) (string, error) {
	var out bytes.Buffer
	if err := json.Indent(&out, input, "", "  "); err != nil {
		return "", err
	}
	return out.String(), nil
}

// Minify strips all insignificant whitespace.
func Minify(input []byte) (string, error) {
	var out bytes.Buffer
	if err := json.Compact(&out, input); err != nil {
		return "", err
	}
	return out.String(), nil
}

// Analyze computes structural statistics.
func Analyze(input []byte) (*AnalyzeResult, error) {
	var v any
	dec := json.NewDecoder(bytes.NewReader(input))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}

	res := &AnalyzeResult{Types: map[string]int{}, ByteSize: len(input)}
	var walk func(node any, depth int)
	walk = func(node any, depth int) {
		if depth > res.MaxDepth {
			res.MaxDepth = depth
		}
		switch t := node.(type) {
		case map[string]any:
			res.Objects++
			res.Types["object"]++
			res.Keys += len(t)
			for _, v := range t {
				walk(v, depth+1)
			}
		case []any:
			res.Arrays++
			res.Types["array"]++
			for _, v := range t {
				walk(v, depth+1)
			}
		case string:
			res.Scalars++
			res.Types["string"]++
		case json.Number:
			res.Scalars++
			res.Types["number"]++
		case bool:
			res.Scalars++
			res.Types["bool"]++
		case nil:
			res.Nulls++
			res.Types["null"]++
		}
	}
	walk(v, 1)

	if obj, ok := v.(map[string]any); ok {
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		res.TopKeys = keys
	}
	return res, nil
}

// Query navigates a document by a dot path, e.g. "users.0.name".
func Query(input []byte, path string) (any, error) {
	var v any
	dec := json.NewDecoder(bytes.NewReader(input))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}

	if path == "" {
		return v, nil
	}

	cur := v
	for _, part := range strings.Split(path, ".") {
		switch node := cur.(type) {
		case map[string]any:
			next, ok := node[part]
			if !ok {
				return nil, fmt.Errorf("key %q not found", part)
			}
			cur = next
		case []any:
			idx := -1
			if _, err := fmt.Sscanf(part, "%d", &idx); err != nil || idx < 0 {
				return nil, fmt.Errorf("expected array index, got %q", part)
			}
			if idx >= len(node) {
				return nil, fmt.Errorf("array index %d out of range (len %d)", idx, len(node))
			}
			cur = node[idx]
		default:
			return nil, fmt.Errorf("cannot descend into %T at %q", cur, part)
		}
	}
	return cur, nil
}
