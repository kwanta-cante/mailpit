FROM golang:alpine as builder

ARG VERSION=dev

COPY . /app

WORKDIR /app

RUN apk add --no-cache git npm && \
npm install && npm run package && \
CGO_ENABLED=0 go build -ldflags "-s -w -X github.com/axllent/mailpit/cmd.Version=${VERSION}" -o /mailpit


FROM alpine:latest

COPY --from=builder /mailpit /mailpit

RUN apk add --no-cache tzdata

ENTRYPOINT ["/mailpit"]
