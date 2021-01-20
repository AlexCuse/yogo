module github.com/alexcuse/yogo/quote-enricher

go 1.15

replace github.com/goinvest/iexcloud/v2 => github.com/alexcuse/iexcloud/v2 v2.13.1-0.20210120113632-5753dac89be4

require (
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/ThreeDotsLabs/watermill-kafka/v2 v2.2.0
	github.com/alexcuse/yogo/common v0.0.0-20210120115608-c223b3136664
	github.com/goinvest/iexcloud/v2 v2.13.0
	github.com/google/uuid v1.1.1
)
