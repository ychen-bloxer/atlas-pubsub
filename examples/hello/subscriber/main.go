package main

// A simple example subscriber app. Listens to messages from a grpc PubSub server.
// You can run the server provided in this examples section.

import (
	"context"
	"flag"
	"log"

	pubsub "github.com/infobloxopen/atlas-pubsub"
	"github.com/infobloxopen/atlas-pubsub/examples/hello"
	pubsubgrpc "github.com/infobloxopen/atlas-pubsub/grpc"
	"google.golang.org/grpc"
)

var url = flag.String("url", ":8080", "the grpc url to the pubsub server")
var topic = flag.String("topic", hello.DefaultTopicName, "the topic to subscribe to")
var subscriptionID = flag.String("subID", hello.DefaultSubscriberID, "the subscription ID for the topic")
var languageFilter = flag.String("language", "", "if present, will only show messages with metadata tagged for the given language")

func main() {
	flag.Parse()
	log.Printf("subscribing to server at %s with topic %q and subscriptionID %q", *url, *topic, *subscriptionID)
	conn, err := grpc.Dial(*url, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial to grpc server: %v", err)
	}
	s := pubsubgrpc.NewSubscriber(*topic, *subscriptionID, conn)

	md := make(map[string]string)
	if languageFilter != nil && *languageFilter != "" {
		md["language"] = *languageFilter
		log.Printf("Only receiving messages written in %q", *languageFilter)
	}

	c, e := s.Start(context.Background(), pubsub.Filter(md))

	for {
		select {
		case msg, isOpen := <-c:
			if !isOpen {
				log.Println("subscription channel closed")
				return
			}
			greeting := string(msg.Message())
			log.Printf("received message: %q\t%v", greeting, msg.Metadata())
			go func(msg pubsub.Message) {
				if err := msg.Ack(); err != nil {
					log.Printf("failed to ack messageID %q: %v", msg.MessageID(), err)
				}
			}(msg)
		case err := <-e:
			log.Printf("encountered error reading subscription: %v", err)
		}
	}
}
