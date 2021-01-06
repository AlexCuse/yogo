FROM golang:1.15-alpine AS builder

# add git for go modules
# RUN apk update && apk add --no-cache make git gcc libc-dev libsodium-dev zeromq-dev
WORKDIR /temp

COPY . .

RUN go mod download

RUN go build -o ./yogo-scanner .

# Next image - Copy built Go binary into new workspace
FROM alpine

COPY --from=builder /temp/yogo-scanner /.
COPY --from=builder /temp/configuration.toml /.
COPY --from=builder /temp/signals.toml /.

RUN ls /

CMD [ "/yogo-scanner" ]