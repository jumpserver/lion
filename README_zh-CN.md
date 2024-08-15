# Lion 

**简体中文** · [English](./README.md)

## 介绍

该项目使用 Golang 和 Vue 重构了 JumpServer 的 Guacamole 组件，负责 RDP 和 VNC 的连接。 主要基于 [Apache Guacamole](http://guacamole.apache.org/)
开发。

## 配置

启动的配置文件参考[config_example](config_example.yml)

## 构建镜像

```shell
docker build -t jumpserver/lion .
```

## docker启动

```shell
docker run -d --name jms_lion -p 8081:8081 \
-v $(pwd)/data:/opt/lion/data \
-v $(pwd)/config.yml:/opt/lion/config.yml \
jumpserver/lion
```