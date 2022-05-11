#!/bin/sh
#

# 需要先创建 guacd 存放日志的目录
mkdir -p /opt/lion/data/logs/

if [ $LOG_LEVEL ]; then
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

/usr/bin/supervisord