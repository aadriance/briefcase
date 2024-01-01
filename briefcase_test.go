package main

import (
	"io"
	"os"
	"testing"
)

var DIRNAME = "test"

type TestOutput struct {
	origStdOut *os.File
	pipeRead   *os.File
	pipeWrite  *os.File
}

// helpers

func setVarsForTest(t *testing.T) string {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("BRIEFCASE_DIR", dir)
	t.Setenv("BRIEFCASE_DIRNAME", DIRNAME)
	return dir
}

func stealStdOut(t *testing.T) TestOutput {
	var out TestOutput
	out.origStdOut = os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal("INFRA: Failed to open OS pipe to route stdout")
	}

	out.pipeRead = r
	out.pipeWrite = w
	os.Stdout = w
	return out
}

func getStdOut(output *TestOutput) string {
	out, _ := io.ReadAll(output.pipeRead)
	output.pipeWrite.Close()
	return string(out)
}

func restoreStdOut(output *TestOutput) {
	os.Stdout = output.origStdOut
	output.pipeWrite.Close()
}

// tests

func TestUserVars(t *testing.T) {
	dir := setVarsForTest(t)
	tmpdir := getTempDir()
	folder := getBriefcaseDirName()
	if dir != tmpdir.path {
		t.Fatal("Temp dir not derived from environment")
	}

	if tmpdir.envVar != "BRIEFCASE_DIR" {
		t.Fatal("dir came from unexpected enfironment variable")
	}

	if folder != DIRNAME {
		t.Fatal("Breifcase directory did not match environment variable")
	}
}

func TestSetGet(t *testing.T) {
	setVarsForTest(t)
	args := UserArgs{"MyVar", "data"}
	set(args)
	out := stealStdOut(t)
	get(args)
	restoreStdOut(&out)
	data := getStdOut(&out)
	if args.value != data {
		t.Fatal("Failed to set data, got ", data)
	}
}

func TestSetListPurge(t *testing.T) {
	setVarsForTest(t)
	args := UserArgs{"MyVar", "data"}
	set(args)
	args = UserArgs{"MyVar2", "data2"}
	set(args)
	args = UserArgs{"MyVar3", "data3"}
	set(args)
	out := stealStdOut(t)
	list(args)
	restoreStdOut(&out)
	data := getStdOut(&out)
	if data != "MyVar\nMyVar2\nMyVar3\n" {
		t.Fatal("Not all variables present, got: ", data)
	}

	args = UserArgs{"force", ""}
	purge(args)
	out = stealStdOut(t)
	list(args)
	restoreStdOut(&out)
	data = getStdOut(&out)
	if data != "" {
		t.Fatal("Data present after purge: ", data)
	}
}

func TestSetRemoveGet(t *testing.T) {
	setVarsForTest(t)
	args := UserArgs{"MyVar", "data"}
	set(args)
	out := stealStdOut(t)
	get(args)
	restoreStdOut(&out)
	data := getStdOut(&out)
	if args.value != data {
		t.Fatal("Failed to set data, got ", data)
	}

	remove(args)
	out = stealStdOut(t)
	get(args)
	restoreStdOut(&out)
	data = getStdOut(&out)
	if args.value == data {
		t.Fatal("Data still present after remove: ", data)
	}
}
