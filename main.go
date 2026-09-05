package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

const version = "1.0.0"

func main() {
	var (
		pretty  = flag.Bool("pretty", false, "pretty-print with 2-space indent")
		minify  = flag.Bool("minify", false, "strip all whitespace")
		sortFl  = flag.Bool("sort", false, "recursively sort object keys alphabetically")
		merge   = flag.String("merge", "", "deep-merge a second JSON file over the input")
		stats   = flag.Bool("stats", false, "print structural statistics")
		query   = flag.String("q", "", "dot-path query, e.g. users.0.name")
		lineNum = flag.Bool("explain", false, "on error, print line/column of the problem")
		ver     = flag.Bool("version", false, "print version")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "json-doctor %s — validate, inspect and fix JSON\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage: json-doctor [flags] [file.json]   (reads stdin if no file)\n\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  cat data.json | json-doctor --pretty\n")
		fmt.Fprintf(os.Stderr, "  json-doctor --stats config.json\n")
		fmt.Fprintf(os.Stderr, "  json-doctor -q users.0.email data.json\n")
	}

	flag.Parse()

	if *ver {
		fmt.Println("json-doctor", version)
		return
	}

	input, err := readInput(flag.Args())
	if err != nil {
		fatal(err, false)
	}

	switch {
	case *pretty && *minify:
		fatal(fmt.Errorf("--pretty and --minify are mutually exclusive"), false)
	case *sortFl && (*pretty || *minify || *stats || *query != ""):
		fatal(fmt.Errorf("--sort cannot be combined with other modes"), false)
	case *merge != "" && (*pretty || *minify || *stats || *query != ""):
		fatal(fmt.Errorf("--merge cannot be combined with other modes"), false)
	case *merge != "":
		other, err := os.ReadFile(*merge)
		if err != nil {
			fatal(err, false)
		}
		out, err := mergeFiles(input, other)
		if err != nil {
			fatal(err, false)
		}
		fmt.Println(out)
		return
	case *sortFl:
		out, err := SortKeys(input)
		if err != nil {
			fatal(err, false)
		}
		fmt.Println(out)
		return
	case *stats:
		res, err := Analyze(input)
		if err != nil {
			fatal(err, *lineNum)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
	case *query != "":
		val, err := Query(input, *query)
		if err != nil {
			fatal(err, false)
		}
		printValue(val)
	case *minify:
		out, err := Minify(input)
		if err != nil {
			fatal(err, false)
		}
		fmt.Println(out)
	case *pretty:
		line, col, err := Validate(input)
		if err != nil {
			fatalWithPosition(err, line, col, *lineNum)
		}
		out, err := Pretty(input)
		if err != nil {
			fatal(err, false)
		}
		fmt.Println(out)
	default:
		// default mode: validate
		line, col, err := Validate(input)
		if err != nil {
			fatalWithPosition(err, line, col, *lineNum)
		}
		fmt.Println("✓ valid json")
	}
}

func readInput(args []string) ([]byte, error) {
	if len(args) > 0 {
		return os.ReadFile(args[0])
	}
	return io.ReadAll(bufio.NewReader(os.Stdin))
}

func printValue(v any) {
	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(v)
}

func fatal(err error, _ bool) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

func fatalWithPosition(err error, line, col int, explain bool) {
	if explain && line > 0 {
		fmt.Fprintf(os.Stderr, "error: %v (line %d, column %d)\n", err, line, col)
	} else {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
	os.Exit(1)
}
