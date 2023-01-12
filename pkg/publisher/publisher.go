package publisher

import (
	"encoding/json"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

func NewPublisher(client mqtt.Client) *Publisher {
	return &Publisher{
		client: client,
	}
}

type Publisher struct {
	client mqtt.Client
}

func (it *Publisher) Handle(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "please use POST", http.StatusMethodNotAllowed)
		return
	}

	var pr PublishRequest
	err := json.NewDecoder(request.Body).Decode(&pr)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	if token := it.client.Publish(pr.Topic, 0, false, pr.Message); token.Wait() && token.Error() != nil {
		logrus.WithError(token.Error()).Errorf("could not publish message %v on topic %s", pr.Message, pr.Topic)
		http.Error(writer, token.Error().Error(), http.StatusInternalServerError)
	}
}

type PublishRequest struct {
	Topic   string `json:"topic"`
	Message string `json:"message"`
}
