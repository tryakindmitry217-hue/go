// cmd/producer/main.go (local test producer)
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/nats-io/nats.go"
)

var queries = []string{
	"iphone 15", "red dress", "summer shoes", "macbook pro", "kindle",
	"lego star wars", "nike air max", "samsung tv", "playstation 5", "barbie",
}

func main() {
	nc, err := nats.Connect("nats://nats:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	for {
		q := queries[rand.Intn(len(queries))]
		event := map[string]interface{}{
			"query":     q,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"user_id":   fmt.Sprintf("user%d", rand.Intn(1000)),
		}
		data, _ := json.Marshal(event)
		nc.Publish("search.queries", data)
		time.Sleep(10 * time.Millisecond)
	}
}
