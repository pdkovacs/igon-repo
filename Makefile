mkfile_path = $(abspath $(lastword $(MAKEFILE_LIST)))
export BACKEND_SOURCE_HOME = $(dir $(mkfile_path))

ui-bundle-dir = web/dist
ui-bundle = $(ui-bundle-dir)/bundle.js
app       = igo-repo
frontend  = web/frontend/bundle.js
backend   = igo-repo-backend

define buildinfo =
	echo VERSION=0.0.1 > internal/config/buildinfo.txt
	printf "TIME=" >> internal/config/buildinfo.txt
	date --rfc-3339=ns >> internal/config/buildinfo.txt
	printf "COMMIT=" >> internal/config/buildinfo.txt
	git rev-parse HEAD >> internal/config/buildinfo.txt
endef

define build-go =
	$(buildinfo)
	echo "GOOS: ${GOOS} GOARCH: ${GOARCH}"
		env GOOS=${GOOS} GOARCH=${GOARCH} go build -o igo-repo cmd/main.go
endef

.PHONY: clean test run app
clean:
	go clean -testcache
	rm -f igo-repo
# example command line:
#   export LOCAL_GIT_ONLY=yes; export ICONREPO_DB_HOST=postgres; make clean && time make test 2>&1 | tee ~/workspace/logs/icon-repo-test
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
keycloak-init:
	cd deployments/dev/keycloak/; bash build.sh
ui-bundle: $(ui-bundle)
app: $(app)
backend: $(backend)
frontend: $(frontend)
docker: GOOS=linux
docker: GOARCH=amd64
app-docker: $(app)
	cp $(app) deployments/docker/backend/
	deployments/docker/backend/build.sh
backend-docker: $(backend)
	scripts/make.sh build_backend_docker $(app)
frontend-docker: $(frontend)
	scripts/make.sh build_frontend_docker $(ui-bundle) $(ui-bundle-dir)
watch:
	./scripts/watch.sh $(ui-bundle) $(ui-bundle-dir) 2>&1 | tee ~/workspace/logs/igo-repo-watch-log
