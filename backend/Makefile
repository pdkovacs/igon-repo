.PHONY: clean test run
clean:
	go clean -testcache
test:
	go test -race ./... | egrep -v '^[?].*\[no test files\]$$'
test-verbose:
	go test -v -race ./... | egrep -v '^[?].*\[no test files\]$$'
run:
	go run main.go
build: test
	go build -ldflags "\
		-X 'github.com/pdkovacs/igo-repo/backend/pkg/config.Version=0.0.1' \
		-X 'github.com/pdkovacs/igo-repo/backend/pkg/config.BuildUser=$$(id -u -n)' \
		-X 'github.com/pdkovacs/igo-repo/backend/pkg/config.BuildTime=$$(date)' \
		-X 'github.com/pdkovacs/igo-repo/backend/pkg/config.Commit=$$(git rev-parse HEAD)' \
	" -o igo-repo-backend
