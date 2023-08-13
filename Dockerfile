FROM node:16.17.1-bullseye-slim as ui-build
ARG TARGETARCH
ARG NPM_REGISTRY="https://registry.npmmirror.com"
ENV NPM_REGISTY=$NPM_REGISTRY

RUN set -ex \
    && npm config set registry ${NPM_REGISTRY} \
    && yarn config set registry ${NPM_REGISTRY}

WORKDIR /opt/lion/ui
ADD ui/package.json ui/yarn.lock .
RUN --mount=type=cache,target=/usr/local/share/.cache/yarn,sharing=locked,id=lion \
    yarn install

ADD ui .
RUN --mount=type=cache,target=/usr/local/share/.cache/yarn,sharing=locked,id=lion \
    yarn build

FROM golang:1.20-bullseye as stage-build
LABEL stage=stage-build
ARG TARGETARCH

WORKDIR /opt/lion

ADD go.mod go.sum .

ARG GOPROXY=https://goproxy.io
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux

RUN --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download -x

COPY . .
ARG VERSION
ENV VERSION=$VERSION

RUN --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    export GOFlAGS="-X 'main.Buildstamp=`date -u '+%Y-%m-%d %I:%M:%S%p'`'" \
    && export GOFlAGS="${GOFlAGS} -X 'main.Githash=`git rev-parse HEAD`'" \
    && export GOFlAGS="${GOFlAGS} -X 'main.Goversion=`go version`'" \
    && export GOFlAGS="${GOFlAGS} -X 'main.Version=${VERSION}'" \
    && go build -trimpath -x -ldflags "$GOFlAGS" -o lion .

RUN chmod +x entrypoint.sh

FROM jumpserver/guacd:1.5.3
ARG TARGETARCH

USER root
WORKDIR /opt/lion

ARG DEPENDENCIES="                    \
        ca-certificates               \
        curl                          \
        locales                       \
        supervisor                    \
        telnet"

ARG APT_MIRROR=http://mirrors.ustc.edu.cn

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked,id=lion \
    sed -i "s@http://.*.debian.org@${APT_MIRROR}@g" /etc/apt/sources.list \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && apt-get update \
    && apt-get install -y --no-install-recommends ${DEPENDENCIES} \
    && echo "zh_CN.UTF-8" | dpkg-reconfigure locales \
    && sed -i "s@# export @export @g" ~/.bashrc \
    && sed -i "s@# alias @alias @g" ~/.bashrc \
    && rm -rf /var/lib/apt/lists/*

COPY --from=ui-build /opt/lion/ui/dist ui/dist/
COPY --from=stage-build /opt/lion/lion .
COPY --from=stage-build /opt/lion/config_example.yml .
COPY --from=stage-build /opt/lion/entrypoint.sh .
COPY --from=stage-build /opt/lion/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

ENV LANG=zh_CN.UTF-8

EXPOSE 8081
CMD ["./entrypoint.sh"]
