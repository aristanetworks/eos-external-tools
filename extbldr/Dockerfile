# syntax=docker/dockerfile:1.3
FROM rockylinux:9.0.20220720 AS base
RUN dnf install -y epel-release-9* git-2.* jq-1.* \
    openssl-3.* python3-pip-21.* python3-pyyaml-5.* \
    rpmdevtools-9.* sudo-1.*  && \
    dnf install -y mock-3.* && \
    dnf install -y wget-1.21.* && dnf clean all
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
CMD ["bash"]

FROM base as builder
ARG EXTBLDR_ROOT=extbldr
USER root
RUN dnf install -y golang-1.18.* && dnf clean all
USER extbldr-robot
RUN mkdir -p src && mkdir -p bin
WORKDIR /home/extbldr-robot/src
COPY ./${EXTBLDR_ROOT}/go.mod ./
COPY ./${EXTBLDR_ROOT}/go.sum ./
RUN go mod download
COPY ./${EXTBLDR_ROOT}/*.go ./
COPY ./${EXTBLDR_ROOT}/cmd/ cmd/
COPY ./${EXTBLDR_ROOT}/impl/ impl/
COPY ./${EXTBLDR_ROOT}/util/ util/
COPY ./${EXTBLDR_ROOT}/testutil/ testutil/
COPY ./${EXTBLDR_ROOT}/manifest/ manifest/
COPY ./${EXTBLDR_ROOT}/impl/mock_cfg.template /usr/share/mock_cfg.template
RUN go build -o  /home/extbldr-robot/bin/extbldr && \
    go test ./... && \
    GO111MODULE=off go get -u golang.org/x/lint/golint && \
    PATH="$PATH:$HOME/go/bin" golint -set_exit_status ./... && \
    go vet ./... && \
    test -z "$(gofmt -l .)"

FROM base as deploy
COPY --from=builder /home/extbldr-robot/bin/extbldr /usr/bin/
COPY --from=builder /usr/share/mock_cfg.template /usr/share/mock_cfg.template
USER root
RUN mkdir /var/extbldr && \
    chown  extbldr-robot /var/extbldr
USER extbldr-robot
