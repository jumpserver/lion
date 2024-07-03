# Lion 

**English** · [简体中文](./README_zh-CN.md)

## Introduction

This project using Golang and Vue, handling RDP and VNC connections. It is mainly based on [Apache Guacamole](http://guacamole.apache.org/)

## Configuration

Refer to the configuration file for startup[config_example](config_example.yml)

## Build the image

```shell
docker build -t jumpserver/lion .
```

## Docker start

```shell
docker run -d --name jms_lion -p 8081:8081 \
-v $(pwd)/data:/opt/lion/data \
-v $(pwd)/config.yml:/opt/lion/config.yml \
jumpserver/lion
```