package main

type commandlineOptions struct {
	base_dir      string
	test_limit    uint
	selected_test int
	experimental  bool
	run_server    bool
	screenshot    bool
	port          uint
}

type testInstance struct {
	enabled         bool
	wants_filepaths bool
	experimental    bool
	name            string
	data            string
	runner          func(chan []testResult, chan interface{})
}

type filePaths struct {
	html  []string
	php   []string
	css   []string
	img   []string
	dir   []string
	other []string
}
type testResult struct {
	file        string
	linenumber  string
	message     string
	result_type int
}
