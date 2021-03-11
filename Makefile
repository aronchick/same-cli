# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
GOLANG_VERSION ?= 1.15.6
GOPATH ?= $(HOME)/go

# set to -V
VERBOSE ?=

export GO111MODULE = on
export GO = go
export PYTHON = python3
export PRECOMMIT = poetry run pre-commit

BUILD_DIR = same-cli

IMAGE_BUILDER ?= docker
DOCKERFILE ?= Dockerfile
OPERATOR_BINARY_NAME ?= $(shell basename ${PWD})

TAG ?= $(eval TAG := $(shell git describe --tags --always))$(TAG)
REPO ?= $(shell echo $$(cd ../${BUILD_DIR} && git config --get remote.origin.url) | sed 's/git@\(.*\):\(.*\).git$$/https:\/\/\1\/\2/')
BRANCH ?= $(shell cd ../${BUILD_DIR} && git branch | grep '^*' | awk '{print $$2}')
ARCH = linux

RELEASE_USER := SAME-Project
RELEASE_REPO :=SAMPLE-CLI-TESTER

TMPRELEASEWORKINGDIR := $(shell mktemp -d -t same-release-dir.XXXXXXX)
TMPARTIFACTDIR := $(shell mktemp -d -t same-artifact-dir.XXXXXXX)
PACKAGE := $(shell echo "same_$(TAG)_$(ARCH)")

# Basic passphrase for testing
PRIVATE_KEY_PASSPHRASE ?= "PASSPHRASE_FOR_TESTING"

ifdef GITHUB_ACTIONS
PRIVATE_KEY_FILE := private.pem
else
PRIVATE_KEY_FILE := test/testdata/signature_keys/private.pem
endif

ifdef GITHUB_ACTIONS
PUBLIC_KEY_FILE := public.pem
else
PUBLIC_KEY_FILE := test/testdata/signature_keys/public.pem
endif

# Location of junit file
JUNIT_FILE ?= /tmp/report.xml

all: build

# Run go fmt against code
fmt:
	@${GO} fmt ./config ./cmd/... ./pkg/...


# Run go vet against code
vet:
	@${GO} vet ./config ./cmd/... ./pkg/...

################################################################################
# Target: modtidy                                                              #
################################################################################
.PHONY: modtidy
modtidy:
	go mod tidy
	
################################################################################
# Target: check-diff                                                           #
################################################################################
.PHONY: check-diff
check-diff:
	git diff --exit-code ./go.mod # check no changes
	git diff --exit-code ./go.sum # check no changes

## Run all pre-commit hooks
################################################################################
# Target: precommit                                                            #
################################################################################
.PHONY: precommit
precommit:
	${PRECOMMIT} run --all

################################################################################
# Target: build	                                                               #
################################################################################
.PHONY: build
build: build-same

################################################################################
# Target: build-same                                                           #
################################################################################
.PHONY: build-same
build-same: fmt vet
	CGO_ENABLED=0 ARCH=linux GOARCH=amd64 ${GO} build -gcflags '-N -l' -ldflags "-X main.VERSION=$(TAG)" -o bin/$(ARCH)/same main.go
	cp bin/$(ARCH)/same bin/same

# Fast rebuilds useful for development.
# Does not regenerate code; assumes you already ran build-same once.
################################################################################
# Target: build-same-fast                                                      #
################################################################################
.PHONY: build-same-fast
build-same-fast: fmt vet
	CGO_ENABLED=0 ARCH=linux GOARCH=amd64 ${GO} build -gcflags '-N -l' -ldflags "-X main.VERSION=$(TAG)" -o bin/$(ARCH)/same main.go

# Release tarballs suitable for upload to GitHub release pages
################################################################################
# Target: build-same-tgz                                                       #
################################################################################
.PHONY: build-same-tgz
build-same-tgz: build-same
	@echo "CWD: $(shell pwd)"
	@echo "RELEASE DIR: $(TMPRELEASEWORKINGDIR)"
	@echo "ARTIFACT DIR: $(TMPARTIFACTDIR)"
	mkdir $(TMPARTIFACTDIR)/$(PACKAGE)
	cp bin/$(ARCH)/same $(TMPARTIFACTDIR)/$(PACKAGE)/same
	cd $(TMPRELEASEWORKINGDIR)
	openssl dgst -sha256 -sign $(PRIVATE_KEY_FILE) -passin pass:$(PRIVATE_KEY_PASSPHRASE) -out $(TMPRELEASEWORKINGDIR)/sign.sha256 $(TMPARTIFACTDIR)/$(PACKAGE)/same
	openssl base64 -in $(TMPRELEASEWORKINGDIR)/sign.sha256 -out $(TMPARTIFACTDIR)/$(PACKAGE)/same.signature.sha256
	@echo "tar cvzf $(TMPARTIFACTDIR)/$(PACKAGE).tar.gz -C $(TMPARTIFACTDIR)/$(PACKAGE) $(PACKAGE)"
	tar cvzf $(TMPARTIFACTDIR)/$(PACKAGE).tar.gz -C $(TMPARTIFACTDIR)/$(PACKAGE) .
	openssl dgst -sha256 -sign $(PRIVATE_KEY_FILE)  -passin pass:$(PRIVATE_KEY_PASSPHRASE) -out $(TMPRELEASEWORKINGDIR)/tarsign.sha256 $(TMPARTIFACTDIR)/$(PACKAGE).tar.gz
	openssl base64 -in $(TMPRELEASEWORKINGDIR)/tarsign.sha256 -out $(TMPARTIFACTDIR)/$(PACKAGE).tar.gz.signature.sha256
	@echo "BINARY_TARBALL=$(TMPARTIFACTDIR)/$(PACKAGE).tar.gz" >> $(GITHUB_ENV)
	@echo "BINARY_TARBALL_NAME=$(PACKAGE).tar.gz" >> $(GITHUB_ENV)
	@echo "BINARY_TARBALL_SIGNATURE=$(TMPARTIFACTDIR)/$(PACKAGE).tar.gz.signature.sha256" >> $(GITHUB_ENV)
	@echo "BINARY_TARBALL_SIGNATURE_NAME=$(PACKAGE).tar.gz.signature.sha256" >> $(GITHUB_ENV)

################################################################################
# Target: build-same-container                                                 #
################################################################################
.PHONY: build-same-container
build-same-container:
	DOCKER_BUILDKIT=1 docker build \
                --build-arg REPO="$(REPO)" \
                --build-arg BRANCH=$(BRANCH) \
		--build-arg GOLANG_VERSION=$(GOLANG_VERSION) \
		--build-arg VERSION=$(TAG) \
		--target=$(SAME_TARGET) \
		--tag $(SAME_IMG)/builder:$(TAG) .
	@echo Built $(SAME_IMG)/builder:$(TAG)
	mkdir -p bin
	docker create \
		--name=temp_same_container \
		$(SAME_IMG)/builder:$(TAG)
	docker cp temp_same_container:/usr/local/bin/same ./bin/same
	docker rm temp_same_container
	@echo Exported same binary to bin/same

# Build but don't attach the latest tag. This allows manual testing/inspection of the image
# first.
################################################################################
# Target: push                                                                 #
################################################################################
.PHONY: push
push: build
	docker push $(BOOTSTRAPPER_IMG):$(TAG)
	@echo Pushed $(BOOTSTRAPPER_IMG):$(TAG)

################################################################################
# Target: install                                                              #
################################################################################
.PHONY: install
install: build-same dockerfordesktop.so
	@echo copying bin/same to /usr/local/bin
	@cp bin/same /usr/local/bin

#***************************************************************************************************
# Build a docker container that can be used to build same
#
# The rules in this section are used to build the docker image that provides
# a suitable go build environment for same
################################################################################
# Target: build-builder-container                                              #
################################################################################
.PHONY: build-builder-container
build-builder-container:
	docker build \
		--build-arg GOLANG_VERSION=$(GOLANG_VERSION) \
		--target=builder \
		--tag $(SAME_IMG):$(TAG) .
	@echo Built $(SAME_IMG):$(TAG)

#***************************************************************************************************

################################################################################
# Target: clean					                                               #
################################################################################
.PHONY: clean
clean:
	rm -rf test && mkdir test

################################################################################
# Target: check-licenses		                                               #
################################################################################
.PHONY: check-licenses
check-licenses:
	# ./third_party/check-license.sh

#
# rules to run unittests
#

################################################################################
# Target: test					                                               #
################################################################################
.PHONY: test
test: build-same check-licenses
	go test ./test/... -v

################################################################################
# Target: lint					                                               #
################################################################################
.PHONY: lint
lint: build-same
	golangci-lint run

# Run the unittests and output a junit report for use with prow
################################################################################
# Target: test-junit			                                               #
################################################################################
.PHONY: test-junit
test-junit: build-same
	echo Running tests ... junit_file=$(JUNIT_FILE)
	go test ./... -v 2>&1 | go-junit-report > $(JUNIT_FILE) --set-exit-code

#***************************************************************************************************

################################################################################
# Target: test-init 			                                               #
################################################################################
.PHONY: test-init
test-init: clean install dockerfordesktop.init none.init-no-platform

test-generate: test-init dockerfordesktop.generate none.generate

test-apply: test-generate dockerfordesktop.apply none.apply

release:
	echo "Executing 'make release'"
	# NOOP
