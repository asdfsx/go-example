# kafka broker的配置

如果使用docker在本地启动kafka，需要添加特殊设置

```
KAFKA_ADVERTISED_HOST_NAME: 127.0.0.1
```

同时容器export出来的端口只能使用9092

问题：

根据 https://hub.docker.com/r/wurstmeister/kafka 的文档，以及kafka的官方文档
，advertise 应该可以使用下面的配置

```
KAFKA_ADVERTISED_LISTENERS: INSIDE://:9092,OUTSIDE://_{HOSTNAME_COMMAND}:9094
KAFKA_LISTENERS: INSIDE://:9092,OUTSIDE://:9094
KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: INSIDE:PLAINTEXT,OUTSIDE:PLAINTEXT
KAFKA_INTER_BROKER_LISTENER_NAME: INSIDE
```