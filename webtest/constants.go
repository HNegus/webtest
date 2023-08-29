package main

import (
	_ "embed"
	"os"
)

const (
	testFail = iota
	testSuccess
	testInfo
	testWarn
	testErr
)

var TMP_HTML_DIR, _ = os.MkdirTemp("", "__WEBTEST")
var VNU_JAR_BASE, _ = os.MkdirTemp("", "__VNU_JAR")
var VNU_JAR_FILENAME = VNU_JAR_BASE + "vnu.jar"

//go:embed lib/vnu.jar
var VNU_JAR_DATA []byte
