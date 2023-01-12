package subscriber

import (
	"net/http"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/oklog/ulid/v2"
	"github.com/sirupsen/logrus"
)

func NewSubscriber(clientOptions mqtt.ClientOptions) *Subscriber {
	return &Subscriber{
		clientOptions: clientOptions,
	}
}

type Subscriber struct {
	clientOptions mqtt.ClientOptions
}

func (it *Subscriber) Handle(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "please use GET", http.StatusMethodNotAllowed)
		return
	}
	topic := request.URL.Query().Get("topic")
	if topic == "" {
		http.Error(writer, "topic not found", http.StatusNotFound)
		return
	}

	it.clientOptions.ClientID = ulid.MustNew(ulid.Now(), ulid.DefaultEntropy()).String()
	client := mqtt.NewClient(&it.clientOptions)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logrus.WithError(token.Error()).Fatal("could not connect to mqtt broker")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer client.Disconnect(0)

	c := make(chan struct{})
	client.Subscribe(topic, 1, func(client mqtt.Client, message mqtt.Message) {
		defer close(c)
		if _, err := writer.Write(message.Payload()); err != nil {
			logrus.WithError(err).Errorf("could not write payload %v", message.Payload())
			return
		}
		message.Ack()
	})
	select {
	case <-c:
	case <-time.After(10 * time.Second):
		writer.WriteHeader(http.StatusRequestTimeout)
	}
}
