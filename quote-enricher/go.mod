module github.com/alexcuse/yogo/quote-enricher

go 1.15

replace github.com/goinvest/iexcloud/v2 => github.com/alexcuse/iexcloud/v2 v2.13.1-0.20210106125043-90ce0f2e6164

require (
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/ThreeDotsLabs/watermill-kafka/v2 v2.2.0
	github.com/alexcuse/yogo/common v0.0.0-20210107010429-058b8a0d1cef
	github.com/goinvest/iexcloud/v2 v2.13.0
	github.com/google/uuid v1.1.1
	github.com/sirupsen/logrus v1.7.0
)
