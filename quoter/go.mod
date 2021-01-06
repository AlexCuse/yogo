module github.com/alexcuse/yogo/quoter

go 1.15

replace (
	github.com/goinvest/iexcloud/v2 => github.com/alexcuse/iexcloud/v2 v2.13.1-0.20210106125043-90ce0f2e6164
)

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/ThreeDotsLabs/watermill-kafka/v2 v2.2.0
	github.com/goinvest/iexcloud/v2 v2.13.0
	github.com/google/uuid v1.1.1
	github.com/hashicorp/go-multierror v1.1.0
)