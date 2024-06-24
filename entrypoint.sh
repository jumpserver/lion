#!/bin/sh
#

if [ -n "$CORE_HOST" ]; then
    until check ${CORE_HOST}/api/health/; do
        echo "wait for jms_core ${CORE_HOST} ready"
        sleep 2
    done
fi

if [ ! -d "/opt/lion/data/logs" ]; then
    mkdir -p /opt/lion/data/logs
fi

: ${LOG_LEVEL:='ERROR'}

case $LOG_LEVEL in
    "DEBUG")
        level="debug"
        ;;
    "INFO")
        level='info'
        ;;
    "WARN")
        level='warning'
        ;;
    "ERROR" | "FATAL" | "CRITICAL")
        level='error'
        ;;
    *)
        level='error'
        ;;
esac
export GUACD_LOG_LEVEL=$level

echo
date
echo "LION Version $VERSION, more see https://www.jumpserver.org"
echo "Quit the server with CONTROL-C."
echo

exec "$@"