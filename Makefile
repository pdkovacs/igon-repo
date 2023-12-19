mkfile_path = $(abspath $(lastword $(MAKEFILE_LIST)))
export BACKEND_SOURCE_HOME = $(dir $(mkfile_path))

backend = iconrepo-backend

# includes UI in the executable
app = iconrepo
ui-dist = ../iconrepo-ui/dist

test-envs = LOG_LEVEL=debug APP_ENV=development

define buildinfo =
	echo VERSION=0.0.1 > internal/config/buildinfo.txt
	printf "TIME=" >> internal/config/buildinfo.txt
	date +%Y-%m-%dT%H:%M:%S%z >> internal/config/buildinfo.txt
	printf "COMMIT=" >> internal/config/buildinfo.txt
	git rev-parse HEAD >> internal/config/buildinfo.txt
endef

define build-go =
	$(buildinfo)
	echo "GOOS: ${GOOS} GOARCH: ${GOARCH}"
	env GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${1} cmd/main.go
endef

.PHONY: clean test run app
clean:
	go clean -testcache
	rm -f iconrepo
# example command line:
#   export LOCAL_GIT_ONLY=yes; export ICONREPO_DB_HOST=postgres; make clean && time make test 2>&1 | tee ~/workspace/logs/iconrepo-test
test: test-server test-iconservice test-repos test-seq
	go test -parallel 1 -v -timeout 60s ./test/iconservice/...
test-server: $(backend)
	go test -parallel 1 -v -timeout 120s ./test/server/...
test-iconservice: $(iconservice)
test-repos: $(backend)
	go test -parallel 1 -v -timeout 60s ./test/repositories/...
test-seq: $(backend)
	go test -parallel 1 -v -timeout 60s ./test/seq/...
test-single: $(backend) # a sample test-case is used, replace it with whichever other test cases you need to run
	go test -parallel 1 -v -timeout 10s ./... -run '^TestAuthBackDoorTestSuite$$' -testify.m TestBackDoorMustntBeAvailableByDefault
test-dynamodb: export DYNAMODB_ONLY = yes
test-dynamodb: export AWS_REGION = eu-west-1
test-dynamodb: backend
	$(test-envs) go test -parallel 1 -v -timeout 40s ./test/repositories/indexing/...
		# -run TestAddTagTestSuite -testify.m TestReuseExistingTag
		# -run TestAddIconToIndexTestSuite -testify.m TestAddASecondIconToIndex
		# -run TestAddIconfileToIndexTestSuite -testify.m TestSecondIconfile
		# -run '(TestAddIconToIndexTestSuite|TestAddIconfileToIndexTestSuite|TestAddTagTestSuite|TestDeleteIconFromIndexTestSuite|TestDeleteIconfileFromIndexTestSuite)'
		# -run TestDeleteIconfileFromIndexTestSuite -testify.m TestRollbackOnFailedSideEffect
run:
	go run cmd/main.go
remove-ui-dist: $(shell find web/dist -type f | grep -v empty.html)
	@rm $^ || echo "No files to remove in web/dist"
$(backend): remove-ui-dist $(shell find internal/ cmd/ -type f)
	# $(call build-go, $(backend))
$(app): $(shell find internal/ cmd/ -type f) $(shell find $(ui-dist) -type f)
	cp -a $(ui-dist)/* web/dist/
	$(call build-go, $(app))
keycloak-init:
	cd deployments/dev/keycloak/; bash build.sh
backend: $(backend)
app: $(app)
docker: GOOS=linux
docker: GOARCH=amd64
backend-docker: $(backend)
	scripts/make.sh build_backend_docker $(backend)
watch:
	./scripts/watch.sh $(ui-bundle) $(ui-bundle-dir) 2>&1 | tee ~/workspace/logs/iconrepo-watch-log
