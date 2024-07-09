FROM node:16.20-bullseye-slim AS ui-build
ARG TARGETARCH
ARG NPM_REGISTRY="https://registry.npmmirror.com"
ENV NPM_REGISTY=$NPM_REGISTRY

RUN set -ex \
    && npm config set registry ${NPM_REGISTRY} \
    && yarn config set registry ${NPM_REGISTRY}

WORKDIR /opt/lion/ui

RUN --mount=type=cache,target=/usr/local/share/.cache/yarn,sharing=locked,id=lion \
    --mount=type=bind,source=ui/package.json,target=package.json \
    --mount=type=bind,source=ui/yarn.lock,target=yarn.lock \
    yarn install

ADD ui .
RUN --mount=type=cache,target=/usr/local/share/.cache/yarn,sharing=locked,id=lion \
    yarn build

FROM golang:1.22-bullseye AS stage-build
LABEL stage=stage-build
ARG TARGETARCH

WORKDIR /opt

ARG CHECK_VERSION=v1.0.2
RUN set -ex \
    && wget https://github.com/jumpserver-dev/healthcheck/releases/download/${CHECK_VERSION}/check-${CHECK_VERSION}-linux-${TARGETARCH}.tar.gz \
    && tar -xf check-${CHECK_VERSION}-linux-${TARGETARCH}.tar.gz \
    && mv check /usr/local/bin/ \
    && chown root:root /usr/local/bin/check \
    && chmod 755 /usr/local/bin/check \
    && rm -f check-${CHECK_VERSION}-linux-${TARGETARCH}.tar.gz

ARG GOPROXY=https://goproxy.io
ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /opt/lion

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked,id=lion \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download

COPY . .

ARG VERSION
ENV VERSION=$VERSION

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked,id=lion \
    export GOFlAGS="-X 'main.Buildstamp=`date -u '+%Y-%m-%d %I:%M:%S%p'`'" \
    && export GOFlAGS="${GOFlAGS} -X 'main.Githash=`git rev-parse HEAD`'" \
    && export GOFlAGS="${GOFlAGS} -X 'main.Goversion=`go version`'" \
    && export GOFlAGS="${GOFlAGS} -X 'main.Version=${VERSION}'" \
    && go build -trimpath -x -ldflags "$GOFlAGS" -o lion .

RUN chmod +x entrypoint.sh

FROM jumpserver/guacd:1.5.5-bullseye
ARG TARGETARCH
ENV LANG=en_US.UTF-8

USER root

ARG DEPENDENCIES="                    \
        ca-certificates               \
        supervisor"

ARG APT_MIRROR=http://mirrors.ustc.edu.cn
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked,id=lion \
    --mount=type=cache,target=/var/lib/apt,sharing=locked,id=lion \
    set -ex \
    && rm -f /etc/apt/apt.conf.d/docker-clean \
    && echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' >/etc/apt/apt.conf.d/keep-cache \
    && sed -i "s@http://.*.debian.org@${APT_MIRROR}@g" /etc/apt/sources.list \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && apt-get update \
    && apt-get install -y --no-install-recommends ${DEPENDENCIES} \
    && sed -i "s@# export @export @g" ~/.bashrc \
    && sed -i "s@# alias @alias @g" ~/.bashrc

WORKDIR /opt/lion

COPY --from=ui-build /opt/lion/ui/dist ui/dist/
COPY --from=stage-build /usr/local/bin /usr/local/bin
COPY --from=stage-build /opt/lion/lion .
COPY --from=stage-build /opt/lion/config_example.yml .
COPY --from=stage-build /opt/lion/entrypoint.sh .
COPY --from=stage-build /opt/lion/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

ARG VERSION
ENV VERSION=$VERSION

VOLUME /opt/lion/data

ENTRYPOINT ["./entrypoint.sh"]

EXPOSE 8081

STOPSIGNAL SIGQUIT

CMD [ "supervisord", "-c", "/etc/supervisor/supervisord.conf" ]