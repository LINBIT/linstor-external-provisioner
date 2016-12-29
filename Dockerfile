# Copyright 2015 bradley childs, All rights reserved.
#

FROM centos:7
MAINTAINER bradley childs, bchilds@gmail.com
# AWS CLI build deps
RUN yum update -y

# Provisioner build deps

# install the shell flex script that the provisioner uses.
RUN mkdir -p  /opt/go/src/github.com/childsb/flex-provision/flex/flex
COPY flex/flex/flex /opt/go/src/github.com/childsb/flex-provision/flex/flex/

# install the go kube piece of provisioner
RUN mkdir -p  /opt/go/src/github.com/childsb/flex-provision/
COPY flex-provision /opt/go/src/github.com/childsb/flex-provision/


ENTRYPOINT ["/opt/go/src/github.com/childsb/flex-provision/flex-provision"]

