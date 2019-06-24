package fileutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
)

func TestCopyDirWithPrefix(t *testing.T) {
	src, err := ioutil.TempDir("", "test-source")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create temp src directory")
	}
	defer os.RemoveAll(src)
	_ = os.MkdirAll(filepath.Join(src, "a"), 0755)
	dst, err := ioutil.TempDir("", "test-dst")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create temp dst directory")
	}
	defer os.RemoveAll(dst)
	srcFiles := []string{"file1.yaml", "file2.yaml", "file3.yaml"}
	for _, f := range srcFiles {
		err = ioutil.WriteFile(filepath.Join(src, "a", f), []byte{}, 0755)
		if err != nil {
			t.Fatalf("failed to create test file %s", f)
		}
	}
	err = CopyDir(filepath.Join(src, "a"), dst, "a")
	if err != nil {
		t.Fatalf("error during copying the directory %v", err)
	}
	fds, err := ioutil.ReadDir(dst)
	var dstFiles []string
	for _, fd := range fds {
		dstFiles = append(dstFiles, fd.Name())
	}
	expFiles := []string{"a-file1.yaml", "a-file2.yaml", "a-file3.yaml"}
	if !cmp.Equal(expFiles, dstFiles) {
		t.Fatalf("expected to get following files %v, but got %v", expFiles, dstFiles)
	}

}

func TestCopyDirWithFiles(t *testing.T) {
	src, err := ioutil.TempDir("", "test-source")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create temp src directory")
	}
	defer os.RemoveAll(src)
	dst, err := ioutil.TempDir("", "test-dst")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create temp dst directory")
	}
	defer os.RemoveAll(dst)
	srcFiles := []string{"file1.yaml", "file2.yaml", "file3.yaml"}
	for _, f := range srcFiles {
		err = ioutil.WriteFile(filepath.Join(src, f), []byte{}, 0755)
		if err != nil {
			t.Fatalf("failed to create test file %s", f)
		}
	}
	err = CopyDir(src, dst, "")
	if err != nil {
		t.Fatal("error during copying the directory")
	}
	fds, err := ioutil.ReadDir(dst)
	var dstFiles []string
	for _, fd := range fds {
		dstFiles = append(dstFiles, fd.Name())
	}

	if !cmp.Equal(srcFiles, dstFiles) {
		t.Fatalf("expected to get following files %v, but got %v", srcFiles, dstFiles)
	}
}
