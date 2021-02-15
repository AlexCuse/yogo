package messaging

import (
	"fmt"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	stan "github.com/nats-io/stan.go"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill-nats/pkg/nats"
	"strings"
)

func NewPublisher(brokerUrl string, clientID string, logger watermill.LoggerAdapter) (message.Publisher, error) {
	cid := fmt.Sprintf("yogo-%s-publisher", clientID)
	if strings.HasPrefix(brokerUrl, "nats") {
		return nats.NewStreamingPublisher(nats.StreamingPublisherConfig{
			ClusterID:   "test-cluster",
			ClientID:    cid,
			StanOptions: []stan.Option{
				stan.NatsURL(brokerUrl),
			},
			Marshaler:   nats.GobMarshaler{},
		}, logger)
	} else {
		saramaConfig := kafka.DefaultSaramaSyncPublisherConfig()
		saramaConfig.ClientID = cid
		return kafka.NewPublisher(kafka.PublisherConfig{
			Brokers:               []string{ brokerUrl },
			Marshaler:             kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: saramaConfig,
		}, logger)
	}
}
