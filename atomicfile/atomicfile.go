package atomicfile

import (
	"fmt"
	"os"
)

type AtomicFile struct {
	fileName     string
	tempFileName string
	tempFile     *os.File
}

const (
	TempSuffix = ".tmp"
)

func Open(filename string) (*AtomicFile, error) {
	tempFileName := fmt.Sprint(filename, TempSuffix)
	tempFile, err := os.Create(tempFileName)
	if err != nil {
		return nil, err
	}

	return &AtomicFile{filename, tempFileName, tempFile}, nil
}

func (a *AtomicFile) Write(p []byte) (n int, err error) {
	if a.tempFile == nil {
		return 0, fmt.Errorf("file is closed")
	}

	return a.tempFile.Write(p)
}

func (a *AtomicFile) Abort() error {
	err := a.Close()
	err2 := os.Remove(a.tempFileName)

	if err2 != nil {
		return err2
	}
	return err
}

func (a *AtomicFile) Commit() error {
	err := a.Close()
	err2 := os.Rename(a.tempFileName, a.fileName)

	if err2 != nil {
		return err2
	}
	return err
}

// Generally you should call Commit or Abort, but Close
// can be used to leave the temp file in place for debugging.
func (a *AtomicFile) Close() error {
	err := a.tempFile.Close()
	a.tempFile = nil

	return err
}
