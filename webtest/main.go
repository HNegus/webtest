package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"
)

func checkCommandAvailable(command string) bool {
	cmd := exec.Command(command, "--version")
	if err := cmd.Run(); err != nil {
		log.Fatal("Error checking command:", command)
		return false
	}
	return true
}

func showAvailableTests(tests []testInstance) {
	printHeading("Available tests")

	for i, test := range tests {
		message := strconv.Itoa(i+1) + ". " + test.name
		if test.experimental {
			message += " [EXPERIMENTAL]"
		}
		if test.enabled {
			printSuccess(message)
		} else {
			printFailure(message)
		}
	}
}

func getExtension(path string) string {

	path = filepath.Ext(path)
	b := []byte(path)
	if match, _ := regexp.Match(`html?`, b); match {
		return "html"
	} else if match, _ := regexp.Match(`(php\d?|pht|phtml)`, b); match {
		return "php"
	} else if match, _ := regexp.Match(`(a?png|p?jpe?g|jfif|pjp|svg|bmp|gif|avif|webp|tiff)`, b); match {
		return "img"
	} else if match, _ := regexp.Match(`css`, b); match {
		return "css"
	} else {
		return "other"

	}
}

func getFilePaths() filePaths {

	dir_queue := []string{"."}

	var result filePaths
	for len(dir_queue) > 0 {

		base_dir := dir_queue[0]
		entries, err := os.ReadDir(base_dir)
		dir_queue = dir_queue[1:]

		if err != nil {
			log.Println("Error reading directory:", err)
			continue
		}

		for _, entry := range entries {

			if entry.IsDir() {
				dir := filepath.Join(base_dir, entry.Name())
				result.dir = append(result.dir, dir)
				dir_queue = append(dir_queue, dir)
				continue
			}

			switch getExtension(entry.Name()) {
			case "html":
				result.html = append(result.html, filepath.Join(base_dir, entry.Name()))
				break
			case "css":
				result.css = append(result.css, filepath.Join(base_dir, entry.Name()))
				continue
			case "php":
				result.php = append(result.php, filepath.Join(base_dir, entry.Name()))
				continue
			case "img":
				result.img = append(result.img, filepath.Join(base_dir, entry.Name()))
				continue
			default:
				result.other = append(result.other, filepath.Join(base_dir, entry.Name()))
				continue
			}
		}
	}
	return result
}

func setup(base_dir string, enable_experimental bool) []testInstance {

	printHeading("Checking dependencies")
	err := os.Chdir(base_dir)
	if err != nil {
		printWarning("Could not find directory: " + base_dir)
	}

	grep_binary_name := absolutePathTestDependencies()

	w3c_validation := validationTestDependencies()

	available_tests := []testInstance{
		{
			name:            "Searching for README.md",
			enabled:         true,
			wants_filepaths: true,
			runner:          runReadmeTest,
		},
		{
			name:            "Searching for homepage",
			enabled:         true,
			wants_filepaths: true,
			runner:          runIndexTest,
		},
		{
			name:            "Searching for large image files",
			enabled:         true,
			wants_filepaths: true,
			runner:          runImageFileSizeTest,
		},
		{
			name:    "Searching for potential absolute paths",
			enabled: grep_binary_name != "",
			runner:  runAbsolutePathTest,
			data:    grep_binary_name,
		},
		{
			name:    "Crawl website for potential dead links",
			enabled: false,
			runner:  runDeadLinkTest,
		},
		{
			name:            "HTML/CSS W3C validation checks",
			enabled:         w3c_validation,
			wants_filepaths: true,
			runner:          runValidationTest,
		},
		{
			name:            "PHP W3C validation checks",
			experimental:    true,
			enabled:         w3c_validation && enable_experimental,
			wants_filepaths: true,
			runner:          runPHPValidationTest,
		},
		{
			name:            "Check for PHP execution errors",
			enabled:         checkCommandAvailable("php") && false,
			wants_filepaths: true,
			runner:          runPHPTest,
		},
	}

	return available_tests
}

func cleanup() {
	os.Remove(VNU_JAR_FILENAME)
	if err := os.RemoveAll(TMP_HTML_DIR); err != nil {
		printWarning(err.Error())
	}
}

func runRoutines(tests []testInstance, config commandlineOptions) {
	filepaths := getFilePaths()

	runTests(tests, config, filepaths)
	printTestsTrailer()

	cleanup()

	if config.run_server {
		runDevSever(filepaths, config.port)
	}

}

func runTests(tests []testInstance, config commandlineOptions, filepaths filePaths) {
	result_channels := make([]chan []testResult, len(tests))
	for i := 0; i < len(tests); i++ {
		result_channels[i] = make(chan []testResult)
	}

	number_of_tests := 0

	for i, test_instance := range tests {
		if test_instance.enabled {
			number_of_tests++
			input_channel := make(chan interface{}, 1)
			go test_instance.runner(result_channels[i], input_channel)

			if test_instance.data != "" {
				input_channel <- test_instance.data
			}

			if test_instance.wants_filepaths {
				input_channel <- filepaths
			}
		}
	}

	for number_of_tests > 0 {
		for i, test_instance := range tests {
			if !test_instance.enabled {
				continue
			}

			select {
			case results := <-result_channels[i]:
				number_of_tests--
				if config.test_limit == 0 || config.test_limit > uint(len(results)) {
					printTestResults(test_instance.name, results)
				} else {
					printTestResults(test_instance.name, results[:config.test_limit])
				}
			default:
				continue
			}
		}
	}
}

func main() {

	var config commandlineOptions
	flag.StringVar(&config.base_dir, "dir", ".", "Specify directory to check.")
	flag.UintVar(&config.test_limit, "l", 0, "Maximum number of results per test. Use 0 for all results.")
	flag.IntVar(&config.selected_test, "test", -1, "Select index of single test to run. Use -1 for all and 0 for no tests.")
	flag.BoolVar(&config.screenshot, "screenshot", false, "[UNAVAILABLE] Make screenshots of all webpages.")
	flag.BoolVar(&config.experimental, "exp", false, "Run experimental tests.")
	flag.BoolVar(&config.run_server, "serve", false, "Run local server for the website.")
	flag.UintVar(&config.port, "p", 8000, "Port to serve on.")

	defer cleanup()

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s\n", os.Args[0])
		flag.PrintDefaults()
		available_tests := setup(config.base_dir, config.experimental)
		showAvailableTests(available_tests)
		fmt.Println()
		cleanup()
	}
	flag.Parse()
	banner()
	available_tests := setup(config.base_dir, config.experimental)
	showAvailableTests(available_tests)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()

	// Disable all tests
	if config.selected_test >= 0 {
		for i := range available_tests {
			available_tests[i].enabled = false
		}
	}
	// Enable only selected test
	if config.selected_test >= 1 && config.selected_test <= len(available_tests) {
		available_tests[config.selected_test-1].enabled = true
	}

	runRoutines(available_tests, config)
}
