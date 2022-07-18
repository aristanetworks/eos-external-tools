# syntax=docker/dockerfile:1.3
FROM quay.io/centos/centos:stream9 AS base
RUN dnf install -y epel-release git jq openssl python3-pip python3-pyyaml rpmdevtools sudo && \
    dnf install -y mock
RUN echo "[jfrog-cli]" > jfrog-cli.repo && \
    echo "name=jfrog-cli" >> jfrog-cli.repo && \ 
    echo "baseurl=https://releases.jfrog.io/artifactory/jfrog-rpms" >> jfrog-cli.repo && \
    echo "enabled=1" >> jfrog-cli.repo && \
    rpm --import https://releases.jfrog.io/artifactory/jfrog-gpg-public/jfrog_public_gpg.key && \
    mv jfrog-cli.repo /etc/yum.repos.d/ && \
    dnf install -y jfrog-cli
RUN useradd -s /bin/bash extbldr-robot -u 10001 -U -p $(openssl passwd -1 extbldr-robot) && \
    useradd -s /bin/bash mockbuild -p $(openssl passwd -1 mockbuild) && \
    usermod -aG mock extbldr-robot
USER extbldr-robot
WORKDIR /home/extbldr-robot
RUN pip3 install wget
CMD bash

FROM base as builder
USER root
RUN dnf install -y golang
USER extbldr-robot
RUN mkdir -p src && mkdir -p bin
WORKDIR /home/extbldr-robot/src
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
COPY cmd/ cmd/
COPY impl/ impl/
COPY util/ util/
RUN go build -o  /home/extbldr-robot/bin/extbldr
RUN go test ./...
RUN GO111MODULE=off go get -u golang.org/x/lint/golint
RUN PATH="$PATH:$HOME/go/bin" golint .
RUN PATH="$PATH:$HOME/go/bin" golint ./...
RUN go vet .
RUN go vet ./...

FROM base as deploy
COPY --from=builder /home/extbldr-robot/bin/extbldr /usr/bin/
