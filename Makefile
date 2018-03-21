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
PROJECT_NAME = `basename $$PWD`
VERSION=`git describe --tags --always --dirty`
LDFLAGS = -ldflags "-X main.Version=${VERSION}"

RM = rm
RM_FLAGS = -vf

all: build

glide:
	glide install --strip-vendor

get:
	-go get ./... &> /dev/null

build: get
	go build $(LDFLAGS)

release: get
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(LDFLAGS) -o $(PROJECT_NAME)-$(OS)-$(ARCH)

clean:
	$(RM) $(RM_FLAGS) $(PROJECT_NAME)

distclean: clean
	$(RM) $(RM_FLAGS) $(PROJECT_NAME)-$(OS)-$(ARCH)
