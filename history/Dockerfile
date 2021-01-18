FROM golang:1.15-alpine AS builder

# add git for go modules
# RUN apk update && apk add --no-cache make git gcc libc-dev libsodium-dev zeromq-dev
WORKDIR /temp

RUN apk add git
ARG GITHUB_KEY
RUN git config --global url."https://golang:$GITHUB_KEY@github.com".insteadOf "https://github.com"

COPY . .

RUN go mod download

RUN go build -o ./yogo-history .

# Next image - Copy built Go binary into new workspace
FROM alpine

COPY --from=builder /temp/yogo-history /.
COPY --from=builder /temp/configuration.toml /.

CMD [ "/yogo-history" ]