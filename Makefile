mkfile_path = $(abspath $(lastword $(MAKEFILE_LIST)))
export BACKEND_SOURCE_HOME = $(dir $(mkfile_path))

ui-bundle-dir = web/dist
ui-bundle = $(ui-bundle-dir)/bundle.js
app       = igo-repo
frontend  = web/frontend/bundle.js
backend   = igo-repo-backend

define build-go =
	echo "GOOS: ${GOOS} GOARCH: ${GOARCH}"
		env GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "\
			-X 'igo-repo/build.version=0.0.1' \
			-X 'igo-repo/build.user=$$(id -u -n)' \
			-X 'igo-repo/build.time=$$(date)' \
			-X 'igo-repo/build.commit=$$(git rev-parse HEAD)' \
		" -o igo-repo cmd/main.go
endef

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
test-single: $(app) # a sample test-case is used, replace it with whichever other test cases you need to run
	go test -parallel 10 -v -timeout 60s ./... -run '^TestIconCreateTestSuite$$' -testify.m TestFailsWith403WithoutPrivilege#01
run:
	go run cmd/main.go
$(ui-bundle): $(shell find web/src -type f) web/webpack.config.js
	cd web; npm install; npm run dist;
$(app): $(ui-bundle) $(shell find internal/ cmd/ -type f)
	$(build-go)
$(backend): $(shell find internal/ cmd/ -type f)
	rm -rf $(ui-bundle-dir); mkdir -p $(ui-bundle-dir); touch $(ui-bundle-dir)/empty.html
	$(build-go)
$(frontend): $(shell find web/src -type f) web/webpack.config.js
	cd web; npm install; npm run frontend;
init-keycloak:
	cd deployments/dev/keycloak/user-client-init
	bash build.sh
ui-bundle: $(ui-bundle)
app: $(app)
backend: $(backend)
frontend: $(frontend)
docker: GOOS=linux
docker: GOARCH=amd64
app-docker: $(app)
	eval "$(minikube docker-env)"
	deployments/docker/build.sh
watch:
	./scripts/watch.sh
