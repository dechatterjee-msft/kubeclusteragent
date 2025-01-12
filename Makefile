
BINARY_NAME ?= kubeclusteragent
CURRENT_VERSION ?= 1.0
BOM_FILE ?= bom.yaml

build-binaries: ## Building the Golang binary
	 GOPATH=$(GOPATH) go mod tidy
	 GOOS=linux GOARCH=amd64 GOPATH=$(GOPATH) go build  -ldflags="-s -w" -o bin/$(BINARY_NAME) cmd/kubeclusteragent/main.go

release-snapshot:
	goreleaser release --snapshot --rm-dist

lint: ## Golang Static code analysis
	go version
	go mod tidy
	@if [ ! -x "`which golangci-lint 2>/dev/null`" ]; then \
		echo "golangci-lint is not found in PATH!!"; \
		exit 1; \
	fi
	golangci-lint --version
	@echo Running code static anaysis...
	golangci-lint run -v
	@echo ""

prepare-mock-env:
	mkdir -p /opt/config
	mkdir -p /var/log/agent
	mkdir -p /etc/containerd
	cp mocks/files/k8s_versions.properties /opt/config/k8s_versions.properties
	cp mocks/files/config.toml /etc/containerd/config.toml


unit-test:prepare-mock-env ## Golang unit testing
	go install github.com/jstemmer/go-junit-report@latest
	go install github.com/t-yuki/gocover-cobertura@latest
	chmod 777 /go/bin/gocover-cobertura
	chmod 777 /go/bin/go-junit-report
	cp  /go/bin/gocover-cobertura /usr/bin/gocover-cobertura
	cp  /go/bin/go-junit-report /usr/bin/go-junit-report
	GOPATH=$(GOPATH) go test -coverprofile=cover.out `go list ./... | grep -v scheduler | grep -v gen | grep -v mocks | grep -v stf` -p 1 -v ./... 2>&1 | go-junit-report -set-exit-code > report.xml
	go tool cover -func cover.out
	gocover-cobertura < cover.out > coverage.xml

unit-local: ## For local environment
	GOPATH=$(GOPATH) go test -coverprofile=cover.out `go list ./... | grep -v scheduler | grep -v gen | grep -v mocks | grep -v stf` -p 1 -v ./... 2>&1 | go-junit-report -set-exit-code > report.xml
	go tool cover -func cover.out | grep "total:"
	go tool cover -html=coverage-report.out -o coverage-report.html


unit-local-docker: prepare-mock-env ## For local environment
	go test -coverprofile=cover.out `go list ./... | grep -v scheduler | grep -v gen | grep -v mocks | grep -v stf` -p 1 -v ./... 2>&1
	go tool cover -func cover.out

build:	## Build
	@echo "Building kubeclusteragent ..."
	export GO111MODULE=off
	go mod tidy
	GOOS=linux GOARCH=$(GOARCH) GOPATH=$(GOPATH) go  build  -ldflags="-s -w" -o bin/$(BINARY_NAME) cmd/kubeclusteragent/main.go
	@echo "Built kubeclusteragent binary"
	@echo "Building kubeclusteragent rpm ..."
	mkdir target
	mkdir ${BINARY_NAME}-${CURRENT_VERSION}
	cp bin/${BINARY_NAME} target/
	mkdir -p target/rpmbuild/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}
	cp bin/${BINARY_NAME} packaging/rpm/${BINARY_NAME}.service ${BINARY_NAME}-${CURRENT_VERSION}
	tar -cvf ${BINARY_NAME}-${CURRENT_VERSION}.tar.gz ${BINARY_NAME}-${CURRENT_VERSION}/
	rm -rf ${BINARY_NAME}-${CURRENT_VERSION}
	mv ${BINARY_NAME}-${CURRENT_VERSION}.tar.gz target/rpmbuild/SOURCES/
	cp packaging/rpm/${BINARY_NAME}.spec target/rpmbuild/SPECS/
	sed -i 's/Version:/Version:        ${CURRENT_VERSION}/' target/rpmbuild/SPECS/${BINARY_NAME}.spec
	rpmbuild --define "_topdir ${PWD}/target/rpmbuild" -bb target/rpmbuild/SPECS/${BINARY_NAME}.spec
	@echo "=========== Successfully built ${BINARY_NAME}-${CURRENT_VERSION} rpm ==========="

setup_docker:
	echo "starting docker"
	sudo /usr/bin/cp support/gobuild/docker/daemon.json /etc/docker/daemon.json
	sudo systemctl restart docker
	sudo /usr/bin/chown mts\:docker /var/run/docker.sock


publish:setup_docker
	mkdir -p ${PUBLISH_DIR}/rpms/
	mkdir -p ${PUBLISH_DIR}/binaries/
	mkdir -p ${PUBLISH_DIR}/scripts
	docker run -v ${PWD}:/kubeclusteragent kubeclusteragent-cicd-photon4:3.2 bash -c "ls -al && cd kubeclusteragent && make build"
	cp target/rpmbuild/RPMS/*/*.rpm ${PUBLISH_DIR}/rpms
	cp target/rpmbuild/BUILD/${BINARY_NAME}-${CURRENT_VERSION}/${BINARY_NAME} ${PUBLISH_DIR}/binaries

local-build-rpm:
	docker run -v ${PWD}:/kubeclusteragent kubeclusteragent-cicd-photon4:3.2 bash -c "ls -al && cd kubeclusteragent && make build"

local-run:
	docker build -f Dockerfile -t kubeclusteragent-cicd-photon4:3.2 .
	docker run -v ${PWD}:/kubeclusteragent kubeclusteragent-cicd-photon4:3.2 bash -c "cd kubeclusteragent && make unit-local-docker"
