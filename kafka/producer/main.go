package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type KafkaProducerOptions struct {
	Addr      string `json:"addr" mapstructure:"addr"`
	Brokers   string `json:"brokers" mapstructure:"brokers"`
	Verbose   bool   `json:"verbose" mapstructure:"verbose"`
	CertFile  string `json:"cert_file" mapstructure:"cert_file"`
	KeyFile   string `json:"key_file" mapstructure:"key_file"`
	CaFile    string `json:"ca_file" mapstructure:"ca_file"`
	VerifySsl bool   `json:"verify_ssl" mapstructure:"verify_ssl"`
}

var defaultKafkaProducerOptions KafkaProducerOptions = KafkaProducerOptions{
	Addr:      ":9095",
	Brokers:   "10.0.0.247:9092,10.0.0.248:9092,10.0.0.230:9092",
	Verbose:   false,
	CertFile:  "",
	KeyFile:   "",
	CaFile:    "",
	VerifySsl: false,
}

func init() {
	flag.StringVar(&defaultKafkaProducerOptions.Addr, "addr", defaultKafkaProducerOptions.Addr, "The address to bind to")
	flag.StringVar(&defaultKafkaProducerOptions.Brokers, "brokers", defaultKafkaProducerOptions.Brokers, "The Kafka brokers to connect to, as a comma separated list")
	flag.BoolVar(&defaultKafkaProducerOptions.Verbose, "verbose", defaultKafkaProducerOptions.Verbose, "Turn on Sarama logging")
	flag.StringVar(&defaultKafkaProducerOptions.CertFile, "certificate", defaultKafkaProducerOptions.CertFile, "The optional certificate file for client authentication")
	flag.StringVar(&defaultKafkaProducerOptions.KeyFile, "key", defaultKafkaProducerOptions.KeyFile, "The optional key file for client authentication")
	flag.StringVar(&defaultKafkaProducerOptions.CaFile, "ca", defaultKafkaProducerOptions.CaFile, "The optional certificate authority file for TLS client authentication")
	flag.BoolVar(&defaultKafkaProducerOptions.VerifySsl, "verify", defaultKafkaProducerOptions.VerifySsl, "Optional verify ssl certificates chain")
	flag.Parse()
	if len(defaultKafkaProducerOptions.Brokers) == 0 {
		panic("no Kafka bootstrap brokers defined, please set the -brokers flag")
	}
}

func main() {
	if defaultKafkaProducerOptions.Verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	brokeList := strings.Split(defaultKafkaProducerOptions.Brokers, ",")

	server := &Server{
		DataCollector:     newDataCollector(brokeList),
		AccessLogProducer: newAccessLogProducer(brokeList),
	}

	defer func() {
		if err := server.Close(); err != nil {
			log.Println("Failed to close server", err)
		}
	}()

	log.Fatal(server.Run(defaultKafkaProducerOptions.Addr))
}

type Server struct {
	DataCollector     sarama.SyncProducer
	AccessLogProducer sarama.AsyncProducer
}

func (s *Server) Close() error {
	if err := s.DataCollector.Close(); err != nil {
		log.Println("Failed to shut down data collector cleanly", err)
	}
	if err := s.AccessLogProducer.Close(); err != nil {
		log.Println("Failed to shut down access log producer cleanly", err)
	}
	return nil
}

func (s *Server) Handler() http.Handler {
	return s.withAccessLog(s.collectQueryStringData())
}

func (s *Server) Run(addr string) error {
	httpServer := &http.Server{
		Addr:    defaultKafkaProducerOptions.Addr,
		Handler: s.Handler(),
	}
	log.Printf("Listening for requests on %s...\n", addr)
	return httpServer.ListenAndServe()
}

func (s *Server) collectQueryStringData() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		// We are not setting a message key, which means that all messages will
		// be distributed randomly over the different partitions.
		partition, offset, err := s.DataCollector.SendMessage(&sarama.ProducerMessage{
			Topic: "important",
			Value: sarama.StringEncoder(r.URL.RawQuery),
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to store your data:, %s", err)
		} else {
			fmt.Fprintf(w, "Your data is stored with unique identifier important/%d/%d", partition, offset)
		}
	})
}

type accessLogEntry struct {
	Method       string  `json:"method"`
	Host         string  `json:"host"`
	Path         string  `json:"path"`
	IP           string  `json:"ip"`
	ResponseTime float64 `json:"response_time"`

	encoded []byte
	err     error
}

func (ale *accessLogEntry) ensureEncoded() {
	if ale.encoded == nil && ale.err == nil {
		ale.encoded, ale.err = json.Marshal(ale)
	}
}

func (ale *accessLogEntry) Length() int {
	ale.ensureEncoded()
	return len(ale.encoded)
}

func (ale *accessLogEntry) Encode() ([]byte, error) {
	ale.ensureEncoded()
	return ale.encoded, ale.err
}

func (s *Server) withAccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		entry := &accessLogEntry{
			Method:       r.Method,
			Host:         r.Host,
			Path:         r.RequestURI,
			IP:           r.RemoteAddr,
			ResponseTime: float64(time.Since(started)) / float64(time.Second),
		}
		// We will use the client's IP address as key. This will cause
		// all the access log entries of the same IP address to end up
		// on the same partition.
		s.AccessLogProducer.Input() <- &sarama.ProducerMessage{
			Topic: "access_log",
			Key:   sarama.StringEncoder(r.RemoteAddr),
			Value: entry,
		}
	})
}

func newDataCollector(brokerList []string) sarama.SyncProducer {
	// For the data collector, we are looking for strong consistency semantics.
	// Because we don't change the flush settings, sarama will try to produce messages
	// as fast as possible to keep latency low.
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   //  Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true
	tlsConfig := createTlsConfiguration()
	if tlsConfig != nil {
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}

	// On the broker side, you may want to change the following settings to get
	// stronger consistency guarantees:
	// - For your broker, set `unclean.leader.election.enable` to false
	// - For the topic, you could increase `min.insync.replicas`.

	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}
	return producer
}

func newAccessLogProducer(brokerList []string) sarama.AsyncProducer {
	// For the access log, we are looking for AP semantics, with high throughput.
	// By creating batches of compressed messages, we reduce network I/O at a cost of more latency.
	config := sarama.NewConfig()
	tlsConfig := createTlsConfiguration()
	if tlsConfig != nil {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}
	config.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy   // Compress messages
	config.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms

	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}

	// We will just log to STDOUT if we're not able to produce messages.
	// Note: messages will only be returned here after all retry attempts are exhausted.
	go func() {
		for err := range producer.Errors() {
			log.Println("Failed to write access log entry:", err)
		}
	}()

	return producer
}

func createTlsConfiguration() (t *tls.Config) {
	if defaultKafkaProducerOptions.CertFile != "" && defaultKafkaProducerOptions.KeyFile != "" &&
		defaultKafkaProducerOptions.CaFile != "" {
		cert, err := tls.LoadX509KeyPair(defaultKafkaProducerOptions.CertFile, defaultKafkaProducerOptions.KeyFile)
		if err != nil {
			log.Fatal(err)
		}
		caCert, err := os.ReadFile(defaultKafkaProducerOptions.CaFile)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		t = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: defaultKafkaProducerOptions.VerifySsl,
		}
	}
	// will be nil by default if nothing is provided
	return
}
