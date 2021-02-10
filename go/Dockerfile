FROM golang:1.15-alpine AS builder

WORKDIR /build

RUN apk add git

COPY go.mod go.sum ./
RUN go mod download

ARG project

COPY internal internal
COPY cmd/$project cmd/$project


RUN go build -o ./service ./cmd/$project/.

# Next image - Copy built Go binary into new workspace
FROM alpine

COPY --from=builder /build/service /.

RUN ls /

CMD [ "/service" ]