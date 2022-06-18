FROM golang:alpine as builder

RUN apk add --no-cache make
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download

COPY ./ .
RUN make dist

FROM alpine:latest

RUN apk add --no-cache ca-certificates
RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /usr/src/app/dist/lsd /usr/bin/lsd
USER app
CMD [ "lsd", "user-service" ]
