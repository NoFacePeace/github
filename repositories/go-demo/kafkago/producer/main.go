package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/segmentio/kafka-go"
)

func main() {
	topic := "pulsar-log-nonlive-test"
	partition := 0
	bootstrapServers := ""
	conn, err := kafka.DialLeader(context.Background(), "tcp", bootstrapServers, topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	defer conn.Close()
	cnt := 0
	for {
		_, err = conn.WriteMessages(
			kafka.Message{Value: []byte(strconv.Itoa(cnt))},
		)
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}
		cnt++
		fmt.Println(cnt)
	}
}
