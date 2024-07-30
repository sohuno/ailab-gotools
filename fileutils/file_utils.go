package fileutils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"time"
)

var (
	ErrFileNotExists = errors.New("file does not exists")
)

func CheckFileExists(file string) bool {
	checkFileFunc := func() (result bool, done bool) {
		defer func() {
			if r := recover(); r != nil {
				klog.Errorf("CheckFileExistsInPanic File:%s Error:%+v", file, r)
				result, done = false, false
			}
		}()
		info, err := os.Stat(file)
		if os.IsNotExist(err) || info == nil {
			return false, true
		}
		return !info.IsDir(), true
	}
	for i := 0; i < 3; i++ {
		if result, done := checkFileFunc(); done {
			return result
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func CheckDirExists(file string) bool {
	checkDirFunc := func() (result bool, done bool) {
		defer func() {
			if r := recover(); r != nil {
				klog.Errorf("CheckDirExistsInPanic File:%s Error:%+v", file, r)
				result, done = false, false
			}
		}()
		info, err := os.Stat(file)
		if os.IsNotExist(err) || info == nil {
			return false, true
		}
		return info.IsDir(), true
	}
	for i := 0; i < 3; i++ {
		if result, done := checkDirFunc(); done {
			return result
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func CheckExists(file string) bool {
	checkExistsFunc := func() (result bool, done bool) {
		defer func() {
			if r := recover(); r != nil {
				klog.Errorf("CheckExistsInPanic File:%s Error:%+v", file, r)
				result, done = false, false
			}
		}()
		_, err := os.Stat(file)
		if os.IsNotExist(err) {
			return false, true
		}
		return true, true
	}
	for i := 0; i < 3; i++ {
		if result, done := checkExistsFunc(); done {
			return result
		}
	}
	return false
}

func CopyFile(src string, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = CopyFileContent(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func CopyFileContent(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func getFileListInternal(path string, flag int) []string {
	var result = make([]string, 0)
	if !CheckDirExists(path) {
		return result
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return result
	}
	for _, file := range files {
		if flag == 0 ||
			(flag == 1 && file.IsDir()) ||
			(flag == 2 && !file.IsDir()) {
			result = append(result, file.Name())
			continue
		}
	}
	return result
}

func GetFileList(path string) []string {
	return getFileListInternal(path, 2)
}

func GetDirList(path string) []string {
	return getFileListInternal(path, 1)
}

func GetAllFiles(path string) []string {
	return getFileListInternal(path, 0)
}

func ReadFileContent(filePath string) (string, error) {
	if bs, err := os.ReadFile(filePath); err != nil {
		return "", err
	} else {
		return string(bs), nil
	}
}

func ReadCgFileLines(filePath string) ([]string, error) {
	return doReadFileAllLines(filePath, 1000)
}

func ReadWorkerExitInfoFile(filePath string) string {
	info, _ := doReadFileAllLines(filePath, 1000)
	if len(info) == 0 {
		return ""
	}
	return info[len(info)-1]
}

func ReadFileLimitedLines(filePath string, limit int) ([]string, error) {
	return doReadFileAllLines(filePath, limit)
}

func ReadFileAllLines(filePath string) ([]string, error) {
	return doReadFileAllLines(filePath, -1)
}

func doReadFileAllLines(filePath string, limit int) ([]string, error) {
	if !CheckFileExists(filePath) {
		return nil, ErrFileNotExists
	}
	readFileAllLines := func(filePath string) (lines []string, err error, done bool) {
		defer func() {
			if r := recover(); r != nil {
				klog.Errorf("ReadFileAllLinesPanic FilePath:%s", filePath)
				lines, err, done = nil, nil, false
			}
		}()
		// open file
		file, err := os.Open(filePath)
		if err != nil {
			klog.Errorf("read file %s failed", filePath)
			return nil, err, true
		}
		// read all lines from filePath
		var results = make([]string, 0)
		{
			var count = 0
			br := bufio.NewReader(file)
			for {
				lineBytes, _, err := br.ReadLine()
				if err == io.EOF {
					break
				}
				results = append(results, string(lineBytes))
				count++
				if limit > 0 && count == limit {
					break
				}
			}
		}
		file.Close()
		return results, nil, true
	}
	var results = make([]string, 0)
	var err error = nil
	var done = false
	for i := 0; i < 3; i++ {
		if results, err, done = readFileAllLines(filePath); done {
			return results, err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return results, err
}

func WriteAllLinesToFile(filePath string, lines []string) error {
	if CheckFileExists(filePath) {
		if err := os.Remove(filePath); err != nil {
			return err
		}
	}
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	bw := bufio.NewWriter(file)
	for _, line := range lines {
		_, _ = bw.WriteString(line + "\n")
	}
	bw.Flush()
	file.Close()
	return nil
}

func WriteFile(filePath string, content string) error {
	if CheckFileExists(filePath) {
		if err := os.Remove(filePath); err != nil {
			return err
		}
	}
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	bw := bufio.NewWriter(file)
	_, err = bw.WriteString(content)
	bw.Flush()
	file.Close()
	return err
}

// AppendFile Append content at the end of the file instead of deleting the original content and rewriting it
func AppendFile(filePath string, content string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func GetDir(path string) string {
	if path == "/" {
		return "/"
	}
	tmp := path
	if strings.HasSuffix(path, "/") {
		tmp = path[0 : len(path)-1]
	}
	pos := strings.LastIndex(tmp, "/")
	if pos == -1 {
		return ""
	}
	return path[0:pos]
}

func UpdateModifiedTime(path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return
	}
	now := time.Now().Local()
	err = os.Chtimes(path, now, now)
	if err != nil {
		klog.Errorf("UpdateModifiedTimeFailed Path:%s Error:%+v", path, err)
	}
}

func ChangeModifiedTime(path string, time time.Time) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return err
	}
	err = os.Chtimes(path, time, time)
	if err != nil {
		klog.Errorf("UpdateModifiedTimeFailed Path:%s Error:%+v", path, err)
		return err
	}
	return nil
}

func GetModifiedTimeInSec(path string) (int64, error) {
	if info, err := os.Stat(path); err == nil {
		return info.ModTime().Unix(), nil
	} else {
		return 0, err
	}
}

func TouchFile(filePath string, content string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	if len(content) > 0 {
		f.WriteString(content)
	}
	f.Close()
	return nil
}
