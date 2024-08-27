# Example Consumer

You can test your event producer fully works by adding a consumer to read the messages published on the Kafka broker; to see it in action, startup your kafka broker (or in our case run `task run-dev`), and then run:

```go
go run pkg/events/kafka/consumer/main.go -brokers="localhost:10000" -topics="sarama" -group="example"
```
