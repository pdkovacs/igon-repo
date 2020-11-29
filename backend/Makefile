.PHONY: clean test run
clean:
	go clean -testcache
test:
	go test -race ./... | egrep -v '^[?].*\[no test files\]$$'
test-verbose:
	go test -v -race ./... | egrep -v '^[?].*\[no test files\]$$'
run:
	go run main.go
