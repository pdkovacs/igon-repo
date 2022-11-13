mkfile_path = $(abspath $(lastword $(MAKEFILE_LIST)))
export BACKEND_SOURCE_HOME = $(dir $(mkfile_path))

ui-bundle = web/dist/bundle.js
app = igo-repo

.PHONY: clean test run app
clean:
	go clean -testcache
	rm -f igo-repo
test: test-app test-api test-repos test-seq
test-app: $(app)
	go test -parallel 10 -v -timeout 60s ./test/app/...
test-api: $(app)
	go test -parallel 10 -v -timeout 120s ./test/api/...
test-repos: $(app)
	go test -parallel 10 -v -timeout 60s ./test/repositories/...
test-seq: $(app)
	go test -parallel 10 -v -timeout 60s ./test/seq/...
test-single: $(app)
	go test -parallel 10 -v -timeout 60s ./... -run '^TestIconCreateTestSuite$$' -testify.m TestFailsWith403WithoutPrivilege#01
run:
	go run cmd/main.go
$(ui-bundle): 
	cd web; npm install; npm run dist;
$(app): $(ui-bundle) $(shell find internal/ cmd/ -type f)
	echo "GOOS: ${GOOS} GOARCH: ${GOARCH}"
	env GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "\
		-X 'igo-repo/build.version=0.0.1' \
		-X 'igo-repo/build.user=$$(id -u -n)' \
		-X 'igo-repo/build.time=$$(date)' \
		-X 'igo-repo/build.commit=$$(git rev-parse HEAD)' \
	" -o igo-repo cmd/main.go
keycloak:
	deployments/dev/keycloak/build.sh
app: $(app)
docker: GOOS=linux
docker: GOARCH=amd64
docker: $(app)
	cp igo-repo deployments/docker
	docker build -t iconrepo:1.0 deployments/docker
watch:
	./scripts/watch.sh
