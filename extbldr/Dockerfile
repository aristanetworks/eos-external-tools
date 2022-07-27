# syntax=docker/dockerfile:1.3
FROM rockylinux:9.0.20220720 AS base
RUN dnf install -y epel-release-9-2.el9 git-2.31.1-2.el9.2 jq-1.6-12.el9 \
    openssl-3.0.1-23.el9_0 python3-pip-21.2.3-6.el9 python3-pyyaml-5.4.1-6.el9 \
    rpmdevtools-9.5-1.el9 sudo-1.9.5p2-7.el9  && \
    dnf install -y mock-3.0-1.el9 && dnf clean all
RUN echo "[jfrog-cli]" > jfrog-cli.repo && \
    echo "name=jfrog-cli" >> jfrog-cli.repo && \ 
    echo "baseurl=https://releases.jfrog.io/artifactory/jfrog-rpms" >> jfrog-cli.repo && \
    echo "enabled=1" >> jfrog-cli.repo && \
    rpm --import https://releases.jfrog.io/artifactory/jfrog-gpg-public/jfrog_public_gpg.key && \
    mv jfrog-cli.repo /etc/yum.repos.d/ && \
    dnf install -y jfrog-cli-1.53.4-1 && dnf clean all
RUN useradd -s /bin/bash extbldr-robot -u 10001 -U -p "$(openssl passwd -1 extbldr-robot)" && \
    useradd -s /bin/bash mockbuild -p "$(openssl passwd -1 mockbuild)" && \
    usermod -aG mock extbldr-robot
USER extbldr-robot
WORKDIR /home/extbldr-robot
RUN pip3 install --no-cache-dir wget==3.2
CMD ["bash"]

FROM base as builder
USER root
RUN dnf install -y golang-1.17.7-1.el9_0 && dnf clean all
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
RUN go build -o  /home/extbldr-robot/bin/extbldr && \
    go test ./... && \
    GO111MODULE=off go get -u golang.org/x/lint/golint && \
    PATH="$PATH:$HOME/go/bin" golint . && \
    PATH="$PATH:$HOME/go/bin" golint ./... && \
    go vet . && \
    go vet ./...

FROM base as deploy
COPY --from=builder /home/extbldr-robot/bin/extbldr /usr/bin/
