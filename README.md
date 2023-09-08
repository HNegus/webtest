# Webtest - Bit Academy

Small tool to automatically test web developer final projects for common mistakes.

Tested on Linux, should also work on MacOS. On Windows it should run using Git Bash.

## Installation

Requires Go version >= 1.19.

### Building from source

Clone the repository first.

```bash
cd webtest
go get && go build
go install
```

Will install the binary in your `$GOPATH`.

## TODO

- [ ] Report errors that occur when trying to execute PHP files
- [ ] Crawl website to search for dead links
- [ ] Check if invalid form submissions are possible
- [ ] Add logging to HTTP server output
- [ ] Auto import SQL if `import.sql` or `*.sql` is found
