package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

type KafkaConsumerOptions struct {
	Brokers  string `json:"brokers" mapstructure:"brokers"`
	Version  string `json:"version" mapstructure:"version"`
	Group    string `json:"group" mapstructure:"group"`
	Topics   string `json:"topics" mapstructure:"topics"`
	Assignor string `json:"assignor" mapstructure:"assignor"`
	Oldest   bool   `json:"oldest" mapstructure:"oldest"`
	Verbose  bool   `json:"verbose" mapstructure:"verbose"`
}

var defaultKafkaConsumerOptions KafkaConsumerOptions = KafkaConsumerOptions{
	Brokers:  "10.0.0.247:9092,10.0.0.248:9092,10.0.0.230:9092",
	Version:  "2.8.1",
	Group:    "sql",
	Topics:   "important,access_log,maxwell",
	Assignor: "range",
	Oldest:   true,
	Verbose:  false,
}

func init() {
	flag.StringVar(&defaultKafkaConsumerOptions.Brokers, "brokers", defaultKafkaConsumerOptions.Brokers, "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&defaultKafkaConsumerOptions.Group, "group", defaultKafkaConsumerOptions.Group, "Kafka consumer group definition")
	flag.StringVar(&defaultKafkaConsumerOptions.Version, "version", defaultKafkaConsumerOptions.Version, "Kafka cluster version")
	flag.StringVar(&defaultKafkaConsumerOptions.Topics, "topics", defaultKafkaConsumerOptions.Topics, "Kafka topics to be consumed, as a comma separated list")
	flag.StringVar(&defaultKafkaConsumerOptions.Assignor, "assignor", defaultKafkaConsumerOptions.Assignor, "Consumer group partition assignment strategy (range, roundrobin, sticky)")
	flag.BoolVar(&defaultKafkaConsumerOptions.Oldest, "oldest", defaultKafkaConsumerOptions.Oldest, "Kafka consumer consume initial offset from oldest")
	flag.BoolVar(&defaultKafkaConsumerOptions.Verbose, "verbose", defaultKafkaConsumerOptions.Verbose, "Sarama logging")
	flag.Parse()

	if len(defaultKafkaConsumerOptions.Brokers) == 0 {
		panic("no Kafka bootstrap brokers defined, please set the -brokers flag")
	}

	if len(defaultKafkaConsumerOptions.Topics) == 0 {
		panic("no topics given to be consumed, please set the -topics flag")
	}

	if len(defaultKafkaConsumerOptions.Group) == 0 {
		panic("no Kafka consumer group defined, please set the -group flag")
	}
}

func main() {
	fmt.Println("kafka configs:")
	prettyJSON, err := json.MarshalIndent(defaultKafkaConsumerOptions, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(prettyJSON))

	if defaultKafkaConsumerOptions.Verbose {
		sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
	}

	version, err := sarama.ParseKafkaVersion(defaultKafkaConsumerOptions.Version)
	if err != nil {
		panic(err)
	}

	config := sarama.NewConfig()
	config.Version = version

	switch defaultKafkaConsumerOptions.Assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	case "roundrobin":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case "range":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	default:
		log.Panicf("Unrecognized consumer group partition assignor: %s", defaultKafkaConsumerOptions.Assignor)
	}

	if defaultKafkaConsumerOptions.Oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	consumer := Consumer{ready: make(chan bool)}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(strings.Split(defaultKafkaConsumerOptions.Brokers, ","), defaultKafkaConsumerOptions.Group, config)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, strings.Split(defaultKafkaConsumerOptions.Topics, ","), &consumer); err != nil {
				fmt.Printf("Error from consumer:%v\n", err)
			}

			if ctx.Err() != nil {
				return
			}

			consumer.ready = make(chan bool)
		}
	}()
	<-consumer.ready
	log.Println("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		log.Println("terminating: context cancelled")
	case <-sigterm:
		log.Println("terminating: via signal")
	}

	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}

}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		log.Printf("Message claimed: key = %s, value = %s, timestamp = %v, topic = %s", string(message.Key), string(message.Value), message.Timestamp, message.Topic)
		session.MarkMessage(message, "")
	}

	return nil
}
