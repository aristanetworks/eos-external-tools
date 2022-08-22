// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package util

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Download the resource srcURL to targetDir
// srcURL could be URL or file path
func Download(srcURL string, targetDir string) (string, error) {
	var uri *url.URL
	uri, parseError := url.ParseRequestURI(srcURL)
	if parseError != nil {
		return "", parseError
	}

	_, statErr := os.Stat(targetDir)
	if statErr != nil {
		return "", statErr
	}

	tokens := strings.Split(uri.Path, "/")
	filename := tokens[len(tokens)-1]

	if uri.Scheme == "file" {
		cpErr := RunSystemCmd("cp", uri.Path, targetDir)
		if cpErr != nil {
			return "", cpErr
		}
	} else {
		if uri.Scheme != "http" && uri.Scheme != "https" {
			return "", fmt.Errorf("util.download: Unsupported URL scheme. (Supported: file, http, https")
		}
		dstPath := filepath.Join(targetDir, filename)

		var file *os.File
		file, createErr := os.Create(dstPath)
		if createErr != nil {
			return "", fmt.Errorf("util.download: Error creating %s", dstPath)
		}
		defer file.Close()

		response, GetErr := http.Get(srcURL)
		if GetErr != nil {
			return "", GetErr
		}

		defer response.Body.Close()
		_, ioErr := io.Copy(file, response.Body)
		if ioErr != nil {
			return "", ioErr
		}
	}
	return filename, nil
}
