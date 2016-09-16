package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update .golden files")

func TestFixFile(t *testing.T) {

	testcases := []struct{ in, golden string }{
		{"simple.input", "simple.golden"},
		{"nested.input", "nested.golden"},
		{"keepcomment.input", "keepcomment.golden"},
	}

	for _, testcase := range testcases {
		var buf bytes.Buffer
		processFile(filepath.Join("testdata", testcase.in), printResult(false, &buf))
		golden := filepath.Join("testdata", testcase.golden)
		if *update {
			ioutil.WriteFile(golden, buf.Bytes(), 0644)
		}

		goldendata, _ := ioutil.ReadFile(golden)

		if !bytes.Equal(buf.Bytes(), goldendata) {
			t.Errorf("want: %q, got %q", string(goldendata), buf.String())
		}
	}

}

func TestFixDir(t *testing.T) {
	testcases := []struct{ in, golden string }{
		{"foo", "foo/all.golden"},
	}

	for _, testcase := range testcases {
		var buf bytes.Buffer
		processDir(filepath.Join("testdata", testcase.in), printResult(false, &buf))
		golden := filepath.Join("testdata", testcase.golden)
		if *update {
			ioutil.WriteFile(golden, buf.Bytes(), 0644)
		}

		goldendata, _ := ioutil.ReadFile(golden)

		if !bytes.Equal(buf.Bytes(), goldendata) {
			t.Errorf("want: %q, got %q", string(goldendata), buf.String())
		}
	}

}
