package main

import (
	_ "embed"
	"os"
	"path/filepath"
)

const (
	testFail = iota
	testSuccess
	testInfo
	testWarn
	testErr
)

var TMP_HTML_DIR, _ = os.MkdirTemp("", "__WEBTEST")

var VNU_JAR_FILENAME = filepath.Join(TMP_HTML_DIR, "vnu.jar")

//go:embed lib/vnu.jar
var VNU_JAR_DATA []byte
