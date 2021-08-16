package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

func isRealtiveFile(path string) bool {
	if filepath.IsAbs(path) {
		return false
	}
	if _, err := url.ParseRequestURI(path); err == nil {
		return false
	}
	return true
}

func hash(src io.Reader) string {
	hash := md5.New()
	if _, err := io.Copy(hash, src); err != nil {
		panic(err)
	}
	return hex.EncodeToString(hash.Sum(nil))[24:]
}

func runCmd(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
