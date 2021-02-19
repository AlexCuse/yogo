package messaging

import (
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill-nats/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	stan "github.com/nats-io/stan.go"
	"strings"
	"time"
)

func NewSubscriber(brokerUrl string, clientID string, logger watermill.LoggerAdapter) (message.Subscriber, error) {
	cid := fmt.Sprintf("yogo-%s-subscriber", clientID)
	if strings.HasPrefix(brokerUrl, "nats://") {
		return nats.NewStreamingSubscriber(nats.StreamingSubscriberConfig{
			ClusterID:        "test-cluster",
			ClientID:         cid,
			QueueGroup:       "yogo",
			SubscribersCount: 50,
			CloseTimeout:     time.Minute,
			AckWaitTimeout:   30 * time.Second,
			StanOptions: []stan.Option{
				stan.NatsURL(brokerUrl),
			},
			StanSubscriptionOptions: [] stan.SubscriptionOption{
				stan.MaxInflight(100_000_000),
			},
			Unmarshaler: nats.GobMarshaler{},
		}, logger)
	} else if strings.HasPrefix(brokerUrl, "amqp") {
		cfg := amqp.NewDurablePubSubConfig(brokerUrl, amqp.GenerateQueueNameTopicNameWithSuffix(cid))
		cfg.Consume.Consumer = cid
		return amqp.NewSubscriber(cfg, logger)
	} else {
		saramaConfig := kafka.DefaultSaramaSubscriberConfig()
		saramaConfig.ClientID = cid
		return kafka.NewSubscriber(kafka.SubscriberConfig{
			Brokers:               []string{brokerUrl},
			Unmarshaler:           kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: saramaConfig,
		}, logger)
	}
}
