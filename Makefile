# Copyright 2017 LINBIT USA LLC.
# Copyright 2016 The Kubernetes Authors.
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

OS=linux
ARCH=amd64

GO = go
PROJECT_NAME = $(shell basename $$PWD)
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS = -X main.Version=${VERSION}
DOCKERREGISTRY=drbd.io

RM = rm
RM_FLAGS = -vf

all: build

glide:
	glide update  --strip-vendor
	glide-vc --only-code --no-tests --use-lock-file

get:
	go get ./... &> /dev/null

build: get
	go build -ldflags '$(LDFLAGS)'

release: get
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -ldflags '$(LDFLAGS)' -o $(PROJECT_NAME)-$(OS)-$(ARCH)

staticrelease: get
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0  $(GO) build -a -ldflags '$(LDFLAGS) -extldflags "-static"' -o $(PROJECT_NAME)-$(OS)-$(ARCH)

dockerimage: distclean
	docker build -t $(DOCKERREGISTRY)/$(PROJECT_NAME) .

clean:
	$(RM) $(RM_FLAGS) $(PROJECT_NAME)-$(OS)-$(ARCH)

distclean: clean
