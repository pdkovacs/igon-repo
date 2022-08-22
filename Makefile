mkfile_path = $(abspath $(lastword $(MAKEFILE_LIST)))
export BACKEND_SOURCE_HOME = $(dir $(mkfile_path))

.PHONY: clean test run build backend
clean:
	go clean -testcache
	rm -f igo-repo
test: ui backend
	go test ./...
test-verbose: backend
	go test -v ./...
test-single: backend
	go test -v ./test/api -run '^TestIconCreateTestSuite$$' -testify.m TestRollbackToLastConsistentStateOnError
run:
	go run cmd/main.go
ui:
	cd web; npm install; npm run dist;
backend:
	echo "GOOS: ${GOOS} GOARCH: ${GOARCH}"
	env GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "\
		-X 'igo-repo/build.version=0.0.1' \
		-X 'igo-repo/build.user=$$(id -u -n)' \
		-X 'igo-repo/build.time=$$(date)' \
		-X 'igo-repo/build.commit=$$(git rev-parse HEAD)' \
	" -o igo-repo cmd/main.go
build: ui backend
docker: GOOS=linux
docker: GOARCH=amd64
docker: build
	cp igo-repo deployments/docker
	docker build -t iconrepo:1.0 deployments/docker
