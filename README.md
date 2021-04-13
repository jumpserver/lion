# Guacamole-client-go

## 介绍

该项目使用 Golang 和 Vue 重构了 JumpServer 的 Guacamole 组件，负责 RDP 和 VNC 的连接。 主要基于 [Apache Guacamole](http://guacamole.apache.org/)
开发。

## 配置

启动的配置文件参考[config_example](config_example.yml)

## 构建镜像

```shell
docker build -t jumpserver/guacamole-client-go .
```

## docker启动

```shell
docker run -d --name jms_guacamole -p 8081:8081 \
-v ./data:/opt/guacamole/data \
-v ./config.yml:/opt/guacamole/config.yml \
jumpserver/guacamole-client-go
```