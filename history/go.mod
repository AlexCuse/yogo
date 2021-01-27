module github.com/alexcuse/yogo/history

go 1.15

replace github.com/goinvest/iexcloud/v2 => github.com/alexcuse/iexcloud/v2 v2.13.1-0.20210124170022-54feaf4794c7

require (
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/ThreeDotsLabs/watermill-kafka/v2 v2.2.0
	github.com/alexcuse/yogo/common v0.0.0-20210126025203-16d51522ef9a
	github.com/alexdrl/zerowater v0.0.3
	github.com/goinvest/iexcloud/v2 v2.13.0
	github.com/rs/zerolog v1.20.0
	github.com/spf13/viper v1.7.1
	gorm.io/datatypes v1.0.0
	gorm.io/driver/postgres v1.0.6
	gorm.io/gorm v1.20.11
)
