module github.com/alexcuse/yogo/scanner

go 1.15

replace github.com/goinvest/iexcloud/v2 => github.com/alexcuse/iexcloud/v2 v2.13.1-0.20210124170022-54feaf4794c7

require (
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/ThreeDotsLabs/watermill-kafka/v2 v2.2.0
	github.com/alexcuse/yogo/common v0.0.0-20210126025203-16d51522ef9a
	github.com/antonmedv/expr v1.8.9
	github.com/goinvest/iexcloud/v2 v2.13.0
	github.com/google/uuid v1.1.1
	github.com/klauspost/compress v1.10.7 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	golang.org/x/sys v0.0.0-20201210223839-7e3030f88018 // indirect
)
