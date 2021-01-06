module github.com/alexcuse/yogo/scanner

go 1.15

replace github.com/goinvest/iexcloud/v2 => github.com/alexcuse/iexcloud/v2 v2.13.1-0.20210106125043-90ce0f2e6164

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/ThreeDotsLabs/watermill-kafka/v2 v2.2.0
	github.com/alexcuse/yogo/common v0.0.0-20210106163145-5e82981a8af6
	github.com/antonmedv/expr v1.8.9
	github.com/goinvest/iexcloud/v2 v2.13.0
	github.com/stretchr/testify v1.6.1
)
