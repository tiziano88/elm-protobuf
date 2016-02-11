package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDiff(t *testing.T) {
	wd, _ := os.Getwd()
	td := filepath.Join(wd, "test")

	dirs, err := ioutil.ReadDir(td)
	if err != nil {
		t.Fatal(err)
	}
	for _, fi := range dirs {
		if !fi.IsDir() {
			continue
		}

		dir := filepath.Join(td, fi.Name())

		err := os.MkdirAll(filepath.Join(dir, "actual_output"), 0777)
		if err != nil {
			t.Fatal(err)
		}

		runProto(t, dir)
		runDiff(t, dir)
	}

}

func runProto(t *testing.T, dir string) {
	cmd := exec.Command("protoc", "--elm_out=../actual_output", "test.proto")
	cmd.Dir = filepath.Join(dir, "input")
	t.Logf("cmd: %v", cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Error: %v, %v", err, string(out))
	}
}

func runDiff(t *testing.T, dir string) {
	cmd := exec.Command("diff", "expected_output", "actual_output")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Error: %v, %v", err, string(out))
	}
}
