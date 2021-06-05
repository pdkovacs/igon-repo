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
		-X 'github.com/pdkovacs/igo-repo/backend/pkg/build.version=0.0.1' \
		-X 'github.com/pdkovacs/igo-repo/backend/pkg/build.user=$$(id -u -n)' \
		-X 'github.com/pdkovacs/igo-repo/backend/pkg/build.time=$$(date)' \
		-X 'github.com/pdkovacs/igo-repo/backend/pkg/build.commit=$$(git rev-parse HEAD)' \
	" -o igo-repo-backend cmd/main.go
