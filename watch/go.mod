module github.com/alexcuse/yogo/watch

go 1.15

replace github.com/goinvest/iexcloud/v2 => github.com/alexcuse/iexcloud/v2 v2.13.1-0.20210119214143-81db86a8c823

require (
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/ThreeDotsLabs/watermill-kafka/v2 v2.2.0
	github.com/alexcuse/yogo/common v0.0.0-20210117212647-e85c83a46ab9
	github.com/gofiber/fiber/v2 v2.3.3
	github.com/goinvest/iexcloud/v2 v2.13.0
	github.com/google/uuid v1.1.1
	github.com/sirupsen/logrus v1.7.0
	gorm.io/driver/postgres v1.0.6
	gorm.io/gorm v1.20.11
)
