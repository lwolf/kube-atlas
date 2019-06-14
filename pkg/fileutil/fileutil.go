package fileutil

import (
	"fmt"
	"io"
	"os"
)

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func CopyFile(src string, dst string) error {
	var fi, dfi os.FileInfo
	var err error
	if fi, err = os.Stat(src); err != nil {
		return err
	}
	if !fi.Mode().IsRegular() {
		return fmt.Errorf("unable to copy non-regular file %s", src)
	}
	if dfi, err = os.Stat(dst); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if !dfi.Mode().IsRegular() {
			return fmt.Errorf("unable to copy to non-regular file %s", dst)
		}
		if os.SameFile(fi, dfi) {
			return fmt.Errorf("unable to copy: same file")
		}
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	err = out.Sync()
	return nil
}
