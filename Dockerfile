# syntax=docker/dockerfile:1.3
FROM rockylinux:9.0.20220720 AS base
RUN dnf install -y epel-release-9* git-2.* jq-1.* \
    openssl-3.* python3-pip-21.* python3-pyyaml-5.* \
    rpmdevtools-9.* sudo-1.*  && \
    dnf install -y mock-3.* automake-1.16.* && \
    dnf install -y wget-1.21.* && dnf clean all
RUN useradd -s /bin/bash eext-robot -u 10001 -U -p "$(openssl passwd -1 eext-robot)" && \
    useradd -s /bin/bash mockbuild -p "$(openssl passwd -1 mockbuild)" && \
    usermod -aG mock eext-robot
USER eext-robot
WORKDIR /home/eext-robot
CMD ["bash"]

FROM base as builder
ARG EEXT_ROOT=.
ARG CFG_DIR=/usr/share/eext
ARG MOCK_CFG_TEMPLATE=mock.cfg.template
ARG REPO_CFG_FILE=dnfrepoconfig.yaml
USER root
RUN dnf install -y golang-1.18.* && dnf clean all
USER eext-robot
RUN mkdir -p src && mkdir -p bin
WORKDIR /home/eext-robot/src
COPY ./${EEXT_ROOT}/go.mod ./
COPY ./${EEXT_ROOT}/go.sum ./
RUN go mod download
COPY ./${EEXT_ROOT}/*.go ./
COPY ./${EEXT_ROOT}/cmd/ cmd/
COPY ./${EEXT_ROOT}/impl/ impl/
COPY ./${EEXT_ROOT}/util/ util/
COPY ./${EEXT_ROOT}/testutil/ testutil/
COPY ./${EEXT_ROOT}/manifest/ manifest/
COPY ./${EEXT_ROOT}/repoconfig/ repoconfig/
COPY ./${EEXT_ROOT}/configfiles/${MOCK_CFG_TEMPLATE} ${CFG_DIR}/${MOCK_CFG_TEMPLATE}
COPY ./${EEXT_ROOT}/configfiles/${REPO_CFG_FILE} ${CFG_DIR}/${REPO_CFG_FILE}
RUN go build -o  /home/eext-robot/bin/eext && \
    go test ./... && \
    GO111MODULE=off go get -u golang.org/x/lint/golint && \
    PATH="$PATH:$HOME/go/bin" golint -set_exit_status ./... && \
    go vet ./... && \
    test -z "$(gofmt -l .)"

FROM base as deploy
COPY --from=builder /home/eext-robot/bin/eext /usr/bin/
COPY --from=builder ${CFG_DIR}/${MOCK_CFG_TEMPLATE} ${CFG_DIR}/${MOCK_CFG_TEMPLATE}
COPY --from=builder ${CFG_DIR}/${REPO_CFG_FILE} ${CFG_DIR}/${REPO_CFG_FILE}
USER root
RUN mkdir /var/eext && \
    chown  eext-robot /var/eext
USER eext-robot
