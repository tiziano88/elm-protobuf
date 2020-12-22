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
	td := filepath.Join(wd, "testdata")

	dirs, err := ioutil.ReadDir(td)
	if err != nil {
		t.Fatal(err)
	}

	for _, fi := range dirs {
		if !fi.IsDir() {
			continue
		}

		dir := filepath.Join(td, fi.Name())
		actualOutputDir := filepath.Join(dir, "actual_output")

		err := os.RemoveAll(actualOutputDir)
		if err != nil {
			t.Fatal(err)
		}

		err = os.MkdirAll(actualOutputDir, 0777)
		if err != nil {
			t.Fatal(err)
		}

		t.Run(fi.Name(), func(t *testing.T) {
			runProto(t, dir)
			runDiff(t, dir)
		})
	}

}

func runProto(t *testing.T, dir string) {
	inputDir := filepath.Join(dir, "input")

	args := []string{"--elm_out=../actual_output"}
	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	for _, file := range files {
		args = append(args, file.Name())
	}

	cmd := exec.Command("protoc", args...)
	cmd.Dir = inputDir
	t.Logf("cmd: %v", cmd)
	out, err := cmd.CombinedOutput()
	t.Logf("Output: %s", out)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func runDiff(t *testing.T, dir string) {
	cmd := exec.Command("diff", "-y", "expected_output", "actual_output")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Error: %v, %v", err, string(out))
	}
}
