#!/bin/sh
#

while [ "$(curl -I -m 10 -L -k -o /dev/null -s -w %{http_code} ${CORE_HOST}/api/health/)" != "200" ]
do
    echo "wait for jms_core ${CORE_HOST} ready"
    sleep 2
done

if [ ! -d "/opt/lion/data/logs" ]; then
    mkdir -p /opt/lion/data/logs
fi

if [ "$LOG_LEVEL" ]; then
        level="info"
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
        esac
        export GUACD_LOG_LEVEL=$level
fi

echo
date
echo "LION Version $VERSION, more see https://www.jumpserver.org"
echo "Quit the server with CONTROL-C."
echo

/usr/bin/supervisord