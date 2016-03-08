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
	td := filepath.Join(wd, "Test")

	dirs, err := ioutil.ReadDir(td)
	if err != nil {
		t.Fatal(err)
	}
	for _, fi := range dirs {
		if !fi.IsDir() {
			continue
		}

		dir := filepath.Join(td, fi.Name())
		actualOutputDir := filepath.Join(dir, "Actual")

		err := os.RemoveAll(actualOutputDir)
		if err != nil {
			t.Fatal(err)
		}

		err = os.MkdirAll(actualOutputDir, 0777)
		if err != nil {
			t.Fatal(err)
		}

		runProto(t, dir)
		runDiff(t, dir)
	}

}

func runProto(t *testing.T, dir string) {
	cmd := exec.Command("protoc", "--elm_out=../Actual", "repeated.proto")
	cmd.Dir = filepath.Join(dir, "Input")
	t.Logf("cmd: %v", cmd)
	out, err := cmd.CombinedOutput()
	t.Logf("Output: %s", out)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func runDiff(t *testing.T, dir string) {
	cmd := exec.Command("diff", "-y", "Expected", "Actual")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Error: %v, %v", err, string(out))
	}
}
