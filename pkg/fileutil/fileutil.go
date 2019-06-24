package fileutil

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

func IsDir(name string) (isDir bool, err error) {
	var fi os.FileInfo
	if fi, err = os.Stat(name); err != nil {
		return false, err
	}
	if fi.Mode().IsDir() {
		return true, nil
	}
	return false, nil
}

// Exists checks whether target file or directory exists
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// CopyDir copies entire directory recursively
func CopyDir(src string, dst string, dstPrefix string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if dstPrefix != "" {
				dstPrefix = fmt.Sprintf("%s-%s", dstPrefix, fd.Name())
			}
			if err = CopyDir(srcfp, dstfp, dstPrefix); err != nil {
				fmt.Println(err)
			}
		} else {
			if dstPrefix != "" {
				dstfp = path.Join(dst, fmt.Sprintf("%s-%s", dstPrefix, fd.Name()))
			}
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// CopyFile tries to copy source file to the destination
func CopyFile(src string, dst string) error {
	var fi, dfi os.FileInfo
	var srcfd *os.File
	var dstfd *os.File
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
	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()
	dstfd, err = os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := dstfd.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if err = dstfd.Sync(); err != nil {
		return err
	}
	if fi, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, fi.Mode())
}
