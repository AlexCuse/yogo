module github.com/alexcuse/yogo/history

go 1.15

replace github.com/goinvest/iexcloud/v2 => github.com/alexcuse/iexcloud/v2 v2.13.1-0.20210124170022-54feaf4794c7

require (
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/ThreeDotsLabs/watermill-kafka/v2 v2.2.0
	github.com/alexcuse/yogo/common v0.0.0-20210123182901-5eac3e03f09d
	github.com/goinvest/iexcloud/v2 v2.13.0
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1 // indirect
	gorm.io/datatypes v1.0.0
	gorm.io/driver/postgres v1.0.6
	gorm.io/gorm v1.20.11
)
