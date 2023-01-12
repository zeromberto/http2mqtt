package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/zeromberto/http2mqtt/pkg/middleware"
	"github.com/zeromberto/http2mqtt/pkg/publisher"
	"github.com/zeromberto/http2mqtt/pkg/subscriber"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type MQTTConfig struct {
	Host             string        `envconfig:"HOST" default:"localhost"`
	Port             int           `envconfig:"PORT" default:"1883"`
	ClientID         string        `envconfig:"CLIENT_ID" default:"http2mqtt"`
	Username         string        `envconfig:"USER"`
	Password         string        `envconfig:"PASSWORD"`
	SubscribeTimeout time.Duration `envconfig:"SUBSCRIBE_TIMEOUT" default:"10s"`
}

type Config struct {
	LogLevel string     `envconfig:"LOG_LEVEL" default:"info"`
	Port     int        `envconfig:"PORT" default:"80"`
	APIKey   string     `envconfig:"API_KEY"`
	MQTT     MQTTConfig `envconfig:"MQTT" required:"true"`
}

func main() {
	cfg := Config{}
	err := envconfig.Process("", &cfg)
	if err != nil {
		logrus.WithError(err).Fatal("invalid config")
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.MQTT.Host, cfg.MQTT.Port))
	opts.SetClientID(cfg.MQTT.ClientID)
	opts.SetUsername(cfg.MQTT.Username)
	opts.SetPassword(cfg.MQTT.Password)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logrus.WithError(token.Error()).Fatal("could not connect to mqtt broker")
	}

	p := publisher.NewPublisher(client)
	s := subscriber.NewSubscriber(*opts, cfg.MQTT.SubscribeTimeout)
	r := mux.NewRouter()
	r.HandleFunc("/publish", p.Handle)
	r.HandleFunc("/subscribe", s.Handle)
	http.Handle("/", middleware.Authenticate(cfg.APIKey, r))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%v", cfg.Port),
		Handler:           r,
		ReadHeaderTimeout: 30 * time.Second,
	}

	logrus.Infof("starting the webserver at %s", server.Addr)
	err = server.ListenAndServe()
	logrus.WithError(err).Error("error while serving bridge")
	client.Disconnect(250)
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	logrus.Errorf("connection lost: %v", err)
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	logrus.Infof("received message: %v from topic: %s", msg.Payload(), msg.Topic())
}
