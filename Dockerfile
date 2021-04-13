FROM node:10 as ui-build
ARG NPM_REGISTRY="https://registry.npm.taobao.org"
ENV NPM_REGISTY=$NPM_REGISTRY

WORKDIR /opt/guacamole
RUN npm config set registry ${NPM_REGISTRY}
RUN yarn config set registry ${NPM_REGISTRY}

COPY ui  ui/
RUN ls . && cd ui/ && npm install -i && yarn build && ls .

# /opt/guacamole/ui/guacamole

FROM golang:1.15-alpine as go-build
WORKDIR /opt/guacamole
ARG GOPROXY=https://goproxy.io
ARG VERSION
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux
COPY go.mod  .
COPY go.sum  .
RUN go mod download
COPY . .
RUN go build -o guacamole-client-go . && ls -al .

FROM guacamole/guacd:1.3.0
USER root
WORKDIR /opt/guacamole
ENV GUACD_LOG_LEVEL=debug
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list \
	&& sed -i 's/security.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list
RUN apt-get update && apt-get install -y supervisor
COPY --from=ui-build /opt/guacamole/ui/guacamole ui/guacamole/
COPY --from=go-build /opt/guacamole/guacamole-client-go .
COPY --from=go-build /opt/guacamole/config_example.yml .
COPY --from=go-build /opt/guacamole/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
RUN ls -al . && pwd
EXPOSE 8081
CMD ["/usr/bin/supervisord"]