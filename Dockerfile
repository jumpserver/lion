FROM node:10 as ui-build
ARG NPM_REGISTRY="https://registry.npmmirror.com"
ENV NPM_REGISTY=$NPM_REGISTRY

WORKDIR /opt/lion
RUN npm config set registry ${NPM_REGISTRY}
RUN yarn config set registry ${NPM_REGISTRY}

COPY ui  ui/
RUN ls . && cd ui/ && npm install -i && yarn build && ls -al .

FROM golang:1.17-alpine as go-build
WORKDIR /opt/lion
ARG GOPROXY=https://goproxy.cn
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux
RUN  sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
     && apk update \
     && apk add git

COPY go.mod  .
COPY go.sum  .
RUN go mod download -x
COPY . .
ARG VERSION=Unknown
RUN export GOFlAGS="-X 'main.Buildstamp=`date -u '+%Y-%m-%d %I:%M:%S%p'`'" \
	&& export GOFlAGS="${GOFlAGS} -X 'main.Githash=`git rev-parse HEAD`'" \
	&& export GOFlAGS="${GOFlAGS} -X 'main.Goversion=`go version`'" \
	&& export GOFlAGS="${GOFlAGS} -X 'main.Version=${VERSION}'" \
	&& go build -trimpath -x -ldflags "$GOFlAGS" -o lion . && ls -al .

FROM jumpserver/guacd:1.4.0
USER root
WORKDIR /opt/lion
RUN sed -i 's@http://.*.debian.org@http://mirrors.ustc.edu.cn@g' /etc/apt/sources.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends supervisor curl telnet iproute2 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=ui-build /opt/lion/ui/lion ui/lion/
COPY --from=go-build /opt/lion/lion .
COPY --from=go-build /opt/lion/config_example.yml .
COPY --from=go-build /opt/lion/entrypoint.sh .
COPY --from=go-build /opt/lion/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
RUN chmod +x entrypoint.sh
CMD ["./entrypoint.sh"]
