# syntax=docker/dockerfile:1.3
FROM rockylinux:9.0.20220720 AS base
RUN dnf install -y epel-release-9* git-2.* jq-1.* \
    openssl-3.* python3-pip-21.* python3-pyyaml-5.* \
    rpmdevtools-9.* sudo-1.*  && \
    dnf install -y mock-3.* automake-1.16.* && \
    dnf install -y wget-1.21.* && dnf clean all
RUN useradd -s /bin/bash lemurbldr-robot -u 10001 -U -p "$(openssl passwd -1 lemurbldr-robot)" && \
    useradd -s /bin/bash mockbuild -p "$(openssl passwd -1 mockbuild)" && \
    usermod -aG mock lemurbldr-robot
USER lemurbldr-robot
WORKDIR /home/lemurbldr-robot
CMD ["bash"]

FROM base as builder
ARG LEMURBLDR_ROOT=lemurbldr
ARG CFG_DIR=/usr/share/lemurbldr
ARG MOCK_CFG_TEMPLATE=mock.cfg.template
ARG REPO_CFG_FILE=dnfrepoconfig.yaml
USER root
RUN dnf install -y golang-1.18.* && dnf clean all
USER lemurbldr-robot
RUN mkdir -p src && mkdir -p bin
WORKDIR /home/lemurbldr-robot/src
COPY ./${LEMURBLDR_ROOT}/go.mod ./
COPY ./${LEMURBLDR_ROOT}/go.sum ./
RUN go mod download
COPY ./${LEMURBLDR_ROOT}/*.go ./
COPY ./${LEMURBLDR_ROOT}/cmd/ cmd/
COPY ./${LEMURBLDR_ROOT}/impl/ impl/
COPY ./${LEMURBLDR_ROOT}/util/ util/
COPY ./${LEMURBLDR_ROOT}/testutil/ testutil/
COPY ./${LEMURBLDR_ROOT}/manifest/ manifest/
COPY ./${LEMURBLDR_ROOT}/repoconfig/ repoconfig/
COPY ./${LEMURBLDR_ROOT}/configfiles/${MOCK_CFG_TEMPLATE} ${CFG_DIR}/${MOCK_CFG_TEMPLATE}
COPY ./${LEMURBLDR_ROOT}/configfiles/${REPO_CFG_FILE} ${CFG_DIR}/${REPO_CFG_FILE}
RUN go build -o  /home/lemurbldr-robot/bin/lemurbldr && \
    go test ./... && \
    GO111MODULE=off go get -u golang.org/x/lint/golint && \
    PATH="$PATH:$HOME/go/bin" golint -set_exit_status ./... && \
    go vet ./... && \
    test -z "$(gofmt -l .)"

FROM base as deploy
COPY --from=builder /home/lemurbldr-robot/bin/lemurbldr /usr/bin/
COPY --from=builder ${CFG_DIR}/${MOCK_CFG_TEMPLATE} ${CFG_DIR}/${MOCK_CFG_TEMPLATE}
COPY --from=builder ${CFG_DIR}/${REPO_CFG_FILE} ${CFG_DIR}/${REPO_CFG_FILE}
USER root
RUN mkdir /var/lemurbldr && \
    chown  lemurbldr-robot /var/lemurbldr
USER lemurbldr-robot
