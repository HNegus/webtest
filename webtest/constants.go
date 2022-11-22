package main

import (
	_ "embed"
	"encoding/base64"
	"os"
)

const (
	testFail = iota
	testSuccess
	testInfo
	testWarn
	testErr
)

var VNU_JAR_FILENAME = base64.StdEncoding.EncodeToString([]byte("vnu.jar"))

var TMP_HTML_DIR, _ = os.MkdirTemp("", "__WEBTEST")

// var TMP_HTML_DIR = "tmpdir"

//go:embed lib/vnu.jar
var VNU_JAR_DATA []byte
