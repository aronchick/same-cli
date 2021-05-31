package utils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func ExecuteCommandC(t *testing.T, root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	// Need to check if we're running in debug mode for VSCode
	// Empty them if they exist
	if (len(os.Args) > 2) && (os.Args[1] == "-test.run") {
		os.Args[1] = ""
		os.Args[2] = ""
	}

	log.Tracef("Command to execute: same %v", root.CalledAs())

	c, err = root.ExecuteC()
	return c, buf.String(), err
}

func PrintErrorAndReturnExit(cmd *cobra.Command, s string, err error) (exit bool) {
	message := fmt.Sprintf(s, err.Error())

	if cmd != nil {
		cmd.Println(message)
		log.Errorf(message)
	}
	return os.Getenv("TEST_PASS") != ""
}

func IsSudoer() bool {
	sudoerCmd := exec.Command("/bin/bash", "-c", "timeout 2 sudo id && echo Access granted || echo Access denied")
	output, err := sudoerCmd.CombinedOutput()
	if strings.Contains(string(output), "Access granted") && err == nil {
		return true
	}

	return false
}

func ClearCobraArgs(cmd *cobra.Command) {
	cmd.SetArgs([]string{})
}

func GetTmpConfigDirectory(runType string) (dirName string) {
	// Create the temp directory
	tmpDirName, err := ioutil.TempDir("", fmt.Sprintf("TEST-SUITE-%v-CONFIG-DIRECTORY-%v", runType, time.Now().UnixNano()))
	if err != nil {
		log.Fatalf("could not create temporary directory to copy files to.")
	}

	return tmpDirName
}

func GetTmpConfigFile(runType string, tmpConfigDirectory string, fileToCopy string) (fileName string, err error) {
	tmpConfigFileHandle, _ := ioutil.TempFile(tmpConfigDirectory, fmt.Sprintf("SAME-TEST-%v-CONFIG-*.yaml", runType))

	return CopyFile(fileToCopy, tmpConfigFileHandle.Name())
}

// File copies a single file from src to dst
func CopyFile(src string, dst string) (string, error) {
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	srcfd, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer func() {
		cerr := srcfd.Close()
		if err == nil {
			err = cerr
		}
	}()

	fileInfo, err := os.Stat(dst)
	if fileInfo != nil && fileInfo.IsDir() {
		log.Tracef("found directory during copy, skipping: %v", fileInfo.Name())
		return "", nil
	} else if os.IsNotExist(err) {
		// Destination file does not exist
		// Using this so we can create a temp file.
		if dstfd, err = os.Create(dst); err != nil {
			return "", err
		}
	} else if err == nil {
		if dstfd, err = os.OpenFile(dst, os.O_APPEND|os.O_WRONLY, os.ModeAppend); err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("could not get information about the file: %v", dst)
	}

	defer func() {
		cerr := dstfd.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return "", err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return "", err
	}
	if err = os.Chmod(dst, srcinfo.Mode()); err != nil {
		return "", err
	}

	return dst, nil
}

// Copies all files in a directory. If dst is empty, creates a directory in tmp.
func CopyFilesInDir(src string, dst string, recursive bool) (string, error) {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return "", err
	}

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		// Destination directory doesn't exist, we'll create it
		// Using this so we can create a temp directory.
		if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
			return "", err
		}
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return "", err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() && recursive {
			if _, err = CopyFilesInDir(srcfp, dstfp, recursive); err != nil {
				return "", err
			}
		} else {
			if fd.IsDir() {
				continue
			}
			if _, err = CopyFile(srcfp, dstfp); err != nil {
				return "", err
			}
		}
	}
	return dst, nil
}
