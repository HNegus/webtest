package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

/************************************************************************
 *								Test runners
 ************************************************************************/
func runValidationTest(result chan []testResult, input chan interface{}) {

	var test_results []testResult
	data := <-input

	filepaths, ok := data.(filePaths)

	if !ok {
		result <- test_results
		return
	}

	test_results = processValidationTest(&filepaths)
	result <- test_results
}

func runPHPValidationTest(result chan []testResult, input chan interface{}) {

	var test_results []testResult
	data := <-input

	filepaths, ok := data.(filePaths)

	if !ok {
		result <- test_results
		return
	}

	test_results = runVnuPHP(&filepaths.php)
	for i := range test_results {
		test_results[i].file = strings.Replace(test_results[i].file, ".html", "", -1)
		test_results[i].file = strings.Replace(test_results[i].file, TMP_HTML_DIR+string(os.PathSeparator), "", -1)
	}
	result <- test_results
}

func runDeadLinkTest(result chan []testResult, input chan interface{}) {

	var test_results []testResult
	result <- test_results
}

func runPHPTest(result chan []testResult, input chan interface{}) {

	var test_results []testResult
	result <- test_results
}

func runIndexTest(result chan []testResult, input chan interface{}) {

	var test_results []testResult
	data := <-input

	var filepaths filePaths
	switch data.(type) {
	case filePaths:
		filepaths = data.(filePaths)
	default:
		result <- test_results
		return
	}

	for _, filename := range filepaths.html {
		if matched, _ := regexp.MatchString("index", strings.ToLower(filename)); matched {
			test_results = append(test_results, testResult{
				file:        filename,
				message:     "Potential homepage",
				result_type: testSuccess,
			})
		}
	}
	for _, filename := range filepaths.php {
		if matched, _ := regexp.MatchString("index", strings.ToLower(filename)); matched {
			test_results = append(test_results, testResult{
				file:        filename,
				message:     "Potential homepage",
				result_type: testSuccess,
			})
		}
	}
	if len(test_results) == 0 {
		result <- []testResult{{
			message:     "No homepage found",
			result_type: testWarn,
		}}
		return
	}
	result <- test_results
}

func runAbsolutePathTest(result chan []testResult, input chan interface{}) {

	data := <-input
	binary_name := ""
	switch data.(type) {
	case string:
		binary_name = data.(string)
	default:
		result <- []testResult{}
		return
	}

	var err error

	output := getAbsolutePathsCommandOutput(binary_name)

	if output == "" {
		result <- []testResult{}
		return
	}

	if err != nil {
		if match, _ := regexp.MatchString("exit status 1", err.Error()); !match {
			log.Println(">> An error occured running "+binary_name, err.Error())
		}
		result <- []testResult{}
		return
	}

	output_string := strings.TrimSpace(string(output))
	result <- processAbsolutePathOutput(output_string)
}

func runImageFileSizeTest(result chan []testResult, input chan interface{}) {

	data := <-input
	var filepaths filePaths
	switch data.(type) {

	case filePaths:
		filepaths = data.(filePaths)
	default:
		result <- []testResult{}
		return

	}
	// filepaths = (data).(filePaths)

	var test_results []testResult

	for _, path := range filepaths.img {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		processed := processImageFile(info)
		if processed.file != "" {
			processed.file = strings.TrimPrefix(path, "./")
			test_results = append(test_results, processed)
		}

	}

	// fmt.Println(filepaths)

	result <- test_results
}

func runReadmeTest(result chan []testResult, input chan interface{}) {
	var test_results []testResult
	data := <-input

	var filepaths filePaths
	switch data.(type) {
	case filePaths:
		filepaths = data.(filePaths)
	default:
		result <- test_results
		return
	}

	for _, filename := range filepaths.other {
		if matched, _ := regexp.MatchString("readme", strings.ToLower(filename)); matched {
			test_results = append(test_results, testResult{
				file:        filename,
				message:     "Potential README",
				result_type: testSuccess,
			})
		}
	}
	if len(test_results) == 0 {
		result <- []testResult{{
			message:     "No README found",
			result_type: testWarn,
		}}
		return
	}
	result <- test_results
}

/************************************************************************
 *								Helper functions
 ************************************************************************/
func processAbsolutePathOutput(output string) []testResult {

	lines := strings.Split(output, "\n")

	result := make([]testResult, len(lines))
	for i, line := range lines {
		parts := strings.SplitAfterN(line, ":", 3)
		result[i].file = strings.TrimSuffix(parts[0], ":")
		result[i].linenumber = strings.TrimSuffix(parts[1], ":")
		result[i].message = strings.TrimSpace(parts[2])
	}

	return result
}

func processImageFile(info os.FileInfo) testResult {

	size := float64(info.Size())

	power := 0
	for ; size/1024 > 1 && power <= 3; power++ {
		size /= 1024
	}

	var result testResult
	suffix := ""

	switch power {
	case 0:
		return testResult{}
	case 1:
		suffix = "Kb"
		if size > 500 {
			result.result_type = testInfo
		} else if size > 750 {
			result.result_type = testWarn
		} else {
			return testResult{}
		}
		break
	case 2:
		suffix = "Mb"
		if size > 1.5 {
			result.result_type = testErr
		} else {
			result.result_type = testFail
		}
		break
	case 3:
		suffix = "Gb"
		result.result_type = testErr
		break
	}

	result.file = info.Name()
	result.message = strconv.FormatFloat(size, 'f', 2, 64) + suffix
	return result
}

func runVnu(filenames *[]string) []testResult {

	if len(*filenames) == 0 {
		return []testResult{}
	}

	params := []string{"-jar", VNU_JAR_FILENAME, "--stdout", "--exit-zero-always"}
	if strings.HasSuffix((*filenames)[0], ".css") {
		params = append(params, "--css")
	}

	params = append(params, *filenames...)

	var cmd *exec.Cmd
	cmd = exec.Command("java", params...)
	output, err := cmd.Output()
	if err != nil {
		return []testResult{}
	}
	return parseVnuResult(string(output))
}

func runVnuPHP(filenames *[]string) []testResult {

	if !checkCommandAvailable("php") || len(*filenames) == 0 {
		return []testResult{}
	}

	pwd, _ := os.Getwd()

	for _, filename := range *filenames {
		new_name := filepath.Join(TMP_HTML_DIR, filename+".html")
		path := filepath.Join(pwd, filepath.Dir(new_name))

		if err := os.MkdirAll(new_name, 0750); err != nil {
			log.Println("Error making file", err)
			continue
		}
		os.Remove(new_name)
		f, err := os.Create(new_name)
		if err != nil {
			log.Println("Error remove file", err)
			continue
		}

		os.Chdir(path)
		cmd := exec.Command("php", "-f", filepath.Join(pwd, filename), new_name)
		cmd.Stdout = f
		if err := cmd.Run(); err != nil {
			// log.Println("Error running php ", err)
		}
	}
	os.Chdir(pwd)

	// Remove line numbers since PHP files are converted to HTML
	test_results := runVnu(&[]string{TMP_HTML_DIR})
	for i := 0; i < len(test_results); i++ {
		test_results[i].linenumber = ""
	}
	return test_results
}

func parseVnuResult(lines string) []testResult {

	var result []testResult
	r := `"file:(?P<file>.*)":(?P<line>.*?): (?P<result_type>.*?): (?P<msg>.*)\.`
	regex := regexp.MustCompile(r)
	lines_split := strings.Split(lines, "\n")

	pwd, _ := os.Getwd()

	for _, line := range lines_split {
		match := regex.FindStringSubmatch(line)

		if len(match) > 0 {
			m := make(map[string]string)
			for i, name := range regex.SubexpNames() {
				if name != "" {
					m[name] = match[i]
				}
			}

			result_type := testInfo
			switch m["result_type"] {
			case "error":
				result_type = testFail
			case "info warning":
				result_type = testWarn
				// case "info":
				// result_type = testInfo
			}

			m["file"] = strings.Replace(m["file"], pwd, "", 1)
			m["file"] = strings.Replace(m["file"], TMP_HTML_DIR, "", 1)
			m["file"] = strings.Replace(m["file"], string(os.PathSeparator), "", 1)

			result = append(result, testResult{
				file:        m["file"],
				linenumber:  m["line"],
				result_type: result_type,
				message:     m["msg"],
			})

		}
	}
	return result
}

func processValidationTest(filepaths *filePaths) []testResult {

	result := runVnu(&filepaths.html)
	result = append(result, runVnu(&filepaths.css)...)

	return result
}

/************************************************************************
 *								Test validators
 ************************************************************************/
func absolutePathTestDependencies() string {

	if checkCommandAvailable("rg") {
		printSuccess("ripgrep available")
		return "rg"
	}
	printWarning("ripgrep not available")

	if checkCommandAvailable("grep") {
		printSuccess("grep available")
		return "grep"
	}
	printFailure("No absolute paths test: grep is not available")
	return ""
}

func validationTestDependencies() bool {

	if !writeVNU() {
		printFailure("No W3C validation tests: Could not build VNU HTML/CSS checker")
		return false
	}

	if !checkCommandAvailable("java") {
		printFailure("No W3C validation tests: java is not available")
		return false
	}

	if !checkCommandAvailable("php") {
		printFailure("No W3C validation tests for PHP files: php is not available")
	} else {
		os.Mkdir(TMP_HTML_DIR, 0750)
		printSuccess("Found php")
		printWarning("W3C validation checks for PHP files are experimental")
	}

	printSuccess("Built VNU HTML/CSS checker")
	printSuccess("Found java")

	return true
}

func writeVNU() bool {
	f, err := os.Create(VNU_JAR_FILENAME)
	defer f.Close()
	if err != nil {
		return false
	}

	f.Write(VNU_JAR_DATA)
	return true
}

func getAbsolutePathsCommandOutput(binary_name string) string {

	var output []byte

	patterns := []string{
		`c:\\`,
		`d:\\`,
		`c:/`,
		`d:/`,
		`/home/`,
		`/Users/`,
	}

	regex := ""
	for _, p := range patterns {
		regex += ".*" + p + ".*|"
	}
	regex = strings.TrimSuffix(regex, "|")
	// fmt.Println(regex)

	var err error

	switch binary_name {
	case "rg":
		output, err = exec.Command("rg", "-inH", "--no-heading", regex).Output()
		// fmt.Println(regex)
	case "grep":
		regex = strings.ReplaceAll(regex, ".*", "")
		// fmt.Println(regex)
		output, err = exec.Command("grep", "-PIRin", regex).Output()
	default:
		break
	}

	if err != nil {
		return string(output)
	}

	return string(output)
}
