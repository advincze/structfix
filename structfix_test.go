package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"testing"
)

var update = flag.Bool("update", false, "update .golden files")

func TestFix(t *testing.T) {
	const (
		input  = "testdata/simple.input"
		golden = "testdata/simple.golden"
	)

	var buf bytes.Buffer
	processFile(input, printResult(false, &buf))

	if *update {
		ioutil.WriteFile(golden, buf.Bytes(), 0644)
	}

	goldendata, _ := ioutil.ReadFile(golden)

	if !bytes.Equal(buf.Bytes(), goldendata) {
		t.Errorf("want: %q, got %q", string(goldendata), buf.String())
	}
}
