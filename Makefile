mkfile_path = $(abspath $(lastword $(MAKEFILE_LIST)))
export BACKEND_SOURCE_HOME = $(dir $(mkfile_path))

.PHONY: clean test run
clean:
	go clean -testcache
test: build
	go test ./...
test-verbose: build
	go test -v ./...
run:
	go run cmd/main.go
build:
	go build -ldflags "\
		-X 'github.com/pdkovacs/igo-repo/internal/build.version=0.0.1' \
		-X 'github.com/pdkovacs/igo-repo/internal/build.user=$$(id -u -n)' \
		-X 'github.com/pdkovacs/igo-repo/internal/build.time=$$(date)' \
		-X 'github.com/pdkovacs/igo-repo/internal/build.commit=$$(git rev-parse HEAD)' \
	" -o igo-repo-backend cmd/main.go
