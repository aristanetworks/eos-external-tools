# syntax=docker/dockerfile:1.3
FROM rockylinux:9.0.20220720 AS base
RUN dnf install -y epel-release-9* git-2.* jq-1.* \
    openssl-3.* python3-pip-21.* python3-pyyaml-5.* \
    rpmdevtools-9.* sudo-1.*  && \
    dnf install -y mock-3.* automake-1.16.* && \
    dnf install -y wget-1.21.* && dnf clean all
RUN useradd -s /bin/bash mockbuild -p "$(openssl passwd -1 mockbuild)"
CMD ["bash"]

FROM base as builder
ARG EEXT_ROOT=.
USER root
RUN dnf install -y golang-1.18.* && dnf clean all
RUN mkdir -p src && mkdir -p bin
WORKDIR /src 
COPY ./${EEXT_ROOT}/go.mod ./
COPY ./${EEXT_ROOT}/go.sum ./
RUN go mod download
COPY ./${EEXT_ROOT}/*.go ./
COPY ./${EEXT_ROOT}/cmd/ cmd/
COPY ./${EEXT_ROOT}/impl/ impl/
COPY ./${EEXT_ROOT}/util/ util/
COPY ./${EEXT_ROOT}/testutil/ testutil/
COPY ./${EEXT_ROOT}/manifest/ manifest/
COPY ./${EEXT_ROOT}/dnfconfig/ dnfconfig/
RUN go build -o  /bin/eext && \
    go test ./... && \
    GO111MODULE=off go get -u golang.org/x/lint/golint && \
    PATH="$PATH:$HOME/go/bin" golint -set_exit_status ./... && \
    go vet ./... && \
    test -z "$(gofmt -l .)"

FROM base as deploy
ARG EEXT_ROOT=.
ARG CFG_DIR=/usr/share/eext
ARG MOCK_CFG_TEMPLATE=mock.cfg.template
ARG DNF_CFG_FILE=dnfconfig.yaml
COPY --from=builder /bin/eext /usr/bin/
COPY ./${EEXT_ROOT}/configfiles/${MOCK_CFG_TEMPLATE} ${CFG_DIR}/${MOCK_CFG_TEMPLATE}
COPY ./${EEXT_ROOT}/configfiles/${DNF_CFG_FILE} ${CFG_DIR}/${DNF_CFG_FILE}
RUN mkdir -p /etc/pki/eext
COPY ./${EEXT_ROOT}/pki/*.pem /etc/pki/eext/
