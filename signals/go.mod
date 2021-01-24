module github.com/alexcuse/yogo/signals

go 1.15

replace github.com/goinvest/iexcloud/v2 => github.com/alexcuse/iexcloud/v2 v2.13.1-0.20210124170022-54feaf4794c7

require (
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/alexcuse/yogo/common v0.0.0-20210123182901-5eac3e03f09d
	github.com/gofiber/fiber/v2 v2.3.3
	github.com/sirupsen/logrus v1.7.0
	gorm.io/driver/postgres v1.0.6
	gorm.io/gorm v1.20.11
)
