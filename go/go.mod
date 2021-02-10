module github.com/alexcuse/yogo

go 1.15

replace github.com/goinvest/iexcloud/v2 => github.com/alexcuse/iexcloud/v2 v2.13.1-0.20210124170022-54feaf4794c7

require (
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/ThreeDotsLabs/watermill-kafka/v2 v2.2.1
	github.com/alexdrl/zerowater v0.0.3
	github.com/antonmedv/expr v1.8.9
	github.com/go-resty/resty/v2 v2.4.0
	github.com/gofiber/fiber/v2 v2.4.1
	github.com/goinvest/iexcloud/v2 v2.13.0
	github.com/google/uuid v1.2.0
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/philippgille/gokv v0.6.0
	github.com/philippgille/gokv/syncmap v0.6.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/zerolog v1.20.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	gorm.io/datatypes v1.0.0
	gorm.io/driver/postgres v1.0.8
	gorm.io/gorm v1.20.12
)
