# syntax=docker/dockerfile:1.3
FROM quay.io/almalinuxorg/9-minimal:9.3-20231124 AS base
RUN microdnf install -y dnf-4.* && microdnf clean all
RUN dnf install -y epel-release-9* git-2.* jq-1.* \
    openssl-3.* python3-pip-21.* python3-pyyaml-5.* \
    rpmdevtools-9.* sudo-1.*  && \
    dnf install -y mock-5.* automake-1.16.* && \
    dnf install -y wget-1.21.* && \
    dnf install -y vim-enhanced-2:8.2.*  emacs-27.* && dnf clean all
RUN useradd -s /bin/bash mockbuild -p "$(openssl passwd -1 mockbuild)"
CMD ["bash"]

FROM base as builder
USER root
RUN dnf install -y golang-1.21.* && dnf clean all
RUN mkdir -p /src/code.arista.io/eos/tools/eext && mkdir -p /usr/bin
WORKDIR /src/code.arista.io/eos/tools/eext
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./*.go ./
COPY ./cmd/ cmd/
COPY ./impl/ impl/
COPY ./util/ util/
COPY ./testutil/ testutil/
COPY ./manifest/ manifest/
COPY ./dnfconfig/ dnfconfig/
COPY ./srcconfig/ srcconfig/
RUN go mod download && go build -o /usr/bin/eext

FROM base as deploy
ARG CFG_DIR=/usr/share/eext
ARG MOCK_CFG_TEMPLATE=mock.cfg.template
ARG DNF_CFG_FILE=dnfconfig.yaml
ARG DNF_CFG_FILE=srcconfig.yaml
COPY --from=builder /bin/eext /usr/bin/
COPY ./configfiles/${MOCK_CFG_TEMPLATE} ${CFG_DIR}/${MOCK_CFG_TEMPLATE}
COPY ./configfiles/${DNF_CFG_FILE} ${CFG_DIR}/${DNF_CFG_FILE}
COPY ./configfiles/${SRC_CFG_FILE} ${CFG_DIR}/${SRC_CFG_FILE}
RUN mkdir -p /etc/pki/eext
COPY ./pki /etc/pki/eext
RUN mkdir /var/eext && chmod 0777 /var/eext && mkdir /dest && chmod 0777 /dest
