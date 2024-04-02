package main

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

func main() {
	group := "pulsar-log-nonlive-test"
	topic := "pulsar-log-nonlive-test"
	bootstrapServers := ""
	// to consume messages
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{bootstrapServers},
		GroupID: group,
		Topic:   topic,
	})
	defer r.Close()
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("message at topic/partition/offset %v/%v/%v/%s: %s = %s\n", m.Topic, m.Partition, m.Offset, m.Time, string(m.Key), string(m.Value))
	}
}
