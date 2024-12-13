# EDA

i want learn EDA

- write functions that use queue to commit
- using golang
- all the infra run locally


## start kafka

`docker-compose up`

## producer and consumer - programmtically

```bash
cd producer
go run producer.go
```

```bash
cd consumer
go run consumer.go
```

## producer and consumer - command

you can add "docker exec -it <kafka_container_id> " to all following command; following command expect you are inside of the conatainer by `docker exec -it <kafka_container_id> bash`.

### creat topics

```
kafka-topics.sh --create --topic test_topic --zookeeper zookeeper:2181 --partitions 1 --replication-factor 1
```

# list topics

```
kafka-topics.sh --list --zookeeper zookeeper:2181 localhost:9092
```

# produce and consume messager

```
kafka-console-producer.sh --topic test_topic --broker-list localhost:9092
```

```
kafka-console-consumer.sh --topic test_topic --bootstrap-server localhost:9092 --from-beginning
```
