package atomicfile_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"zoneupdated/atomicfile"
)

const (
	originalContents = "Hello World\n"
	newContents      = "This is the new contents\n"
)

var (
	filename     string
	tempFilename string
)

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	filename = fmt.Sprintf("%s%ctestfile.%d", os.TempDir(), os.PathSeparator, os.Getpid())
	defer os.Remove(filename)

	tempFilename = filename + atomicfile.TempSuffix
	defer os.Remove(tempFilename)

	return m.Run()
}

func TestAtomicFile_Abort(t *testing.T) {
	createOriginalFile(t, filename)
	a, err := atomicfile.Open(filename)
	if err != nil {
		t.Fatalf("Error opening atomic file: %s", err)
	}

	testWrite(t, a, newContents)
	err = a.Abort()
	if err != nil {
		t.Fatalf("Abort failed: %s", err)
	}

	checkFileMatches(t, filename, originalContents)
	checkFileDoesNotExist(t, tempFilename)
}

func TestAtomicFile_Commit(t *testing.T) {
	createOriginalFile(t, filename)
	a, err := atomicfile.Open(filename)
	if err != nil {
		t.Fatalf("Error opening atomic file: %s", err)
	}

	testWrite(t, a, newContents)
	err = a.Commit()
	if err != nil {
		t.Fatalf("Abort failed: %s", err)
	}

	checkFileMatches(t, filename, newContents)
	checkFileDoesNotExist(t, tempFilename)
}

func TestAtomicFile_Close(t *testing.T) {
	createOriginalFile(t, filename)
	a, err := atomicfile.Open(filename)
	if err != nil {
		t.Fatalf("Error opening atomic file: %s", err)
	}

	testWrite(t, a, newContents)
	err = a.Close()
	if err != nil {
		t.Fatalf("error closing file: %s", err)
	}

	checkFileMatches(t, filename, originalContents)
	checkFileMatches(t, tempFilename, newContents)

	_, err = a.Write([]byte("Try to write after file is closed"))
	if err == nil {
		t.Error("No error when trying to write to closed file")
	}
}

func createOriginalFile(t *testing.T, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Error creating file: %s", err)
	}
	defer f.Close()

	_, _ = f.WriteString(originalContents)
}

func checkFileMatches(t *testing.T, filename string, expectedContents string) {
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Error opening %s: %s", filename, err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("Error reading %s: %s", filename, err)
	}

	if string(data) != expectedContents {
		t.Errorf("Expected file to contain '%s' but contained '%s'",
			expectedContents, string(data))
	}
}

func checkFileDoesNotExist(t *testing.T, filename string) {
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		t.Errorf("File %s exists but is expected not to", filename)
	}
}

func testWrite(t *testing.T, a *atomicfile.AtomicFile, contents string) {
	_, err := a.Write([]byte(contents))
	if err != nil {
		t.Fatalf("Error writing to AtomicFile: %s", err)
	}
}
