# syntax=docker/dockerfile:1
FROM quay.io/centos/centos:stream9 AS base
RUN dnf install -y epel-release git jq openssl python3-pip python3-pyyaml rpmdevtools sudo && dnf install -y mock
RUN echo "[jfrog-cli]" > jfrog-cli.repo && echo "name=jfrog-cli" >> jfrog-cli.repo &&  echo "baseurl=https://releases.jfrog.io/artifactory/jfrog-rpms" >> jfrog-cli.repo && echo "enabled=1" >> jfrog-cli.repo && rpm --import https://releases.jfrog.io/artifactory/jfrog-gpg-public/jfrog_public_gpg.key && mv jfrog-cli.repo /etc/yum.repos.d/ && dnf install -y jfrog-cli
RUN useradd -s /bin/bash arastra -u 10000 -U -p $(openssl passwd -1 arastra) && \
    useradd -s /bin/bash mockbuild -p $(openssl passwd -1 centos) && \
    usermod -aG wheel arastra && \
    usermod -aG mock arastra && \
    echo "arastra ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/arastra
USER arastra
WORKDIR /home/arastra
RUN pip3 install wget
CMD bash

FROM base as builder
RUN sudo dnf install -y golang && mkdir -p src && mkdir -p bin
WORKDIR /home/arastra/src
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
COPY cmd/ cmd/
COPY impl/ impl/
COPY util/ util/
RUN go build -o  /home/arastra/bin/extbldr
RUN go test ./...
RUN go vet .

FROM base as deploy
COPY --from=builder /home/arastra/extbldr /usr/bin/
