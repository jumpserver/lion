FROM jumpserver/guacd:1.5.5-bullseye AS stage-guacd
FROM jumpserver/lion-base:20240719_034830 AS stage-build
ARG TARGETARCH

ARG GOPROXY=https://goproxy.io
ENV CGO_ENABLED=0
ENV GO111MODULE=on

COPY . .

WORKDIR /opt/lion/ui

RUN yarn build

WORKDIR /opt/lion/

ARG VERSION
ENV VERSION=$VERSION

RUN export GOFlAGS="-X 'main.Buildstamp=`date -u '+%Y-%m-%d %I:%M:%S%p'`'" \
    && export GOFlAGS="${GOFlAGS} -X 'main.Githash=`git rev-parse HEAD`'" \
    && export GOFlAGS="${GOFlAGS} -X 'main.Goversion=`go version`'" \
    && export GOFlAGS="${GOFlAGS} -X 'main.Version=${VERSION}'" \
    && go build -trimpath -x -ldflags "$GOFlAGS" -o lion .

RUN chmod +x entrypoint.sh

FROM debian:bullseye-slim
ARG TARGETARCH
ENV LANG=en_US.UTF-8

ARG DEPENDENCIES="                    \
        ca-certificates               \
        supervisor"

ARG PREFIX_DIR=/opt/guacamole
ENV LD_LIBRARY_PATH=${PREFIX_DIR}/lib

COPY --from=stage-guacd ${PREFIX_DIR} ${PREFIX_DIR} 

ARG APT_MIRROR=http://deb.debian.org

RUN set -ex \
    && sed -i "s@http://.*.debian.org@${APT_MIRROR}@g" /etc/apt/sources.list \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && apt-get update \
    && apt-get install -y --no-install-recommends ${DEPENDENCIES} \
    && apt-get install -y --no-install-recommends $(cat "${PREFIX_DIR}"/DEPENDENCIES) \
    && apt-get clean all \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /lib32 /libx32

WORKDIR /opt/lion

COPY --from=stage-build /usr/local/bin/check /usr/local/bin/check
COPY --from=stage-build /opt/lion/ui/dist ui/dist/
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
