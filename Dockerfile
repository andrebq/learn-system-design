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

COPY --from=builder /usr/src/app/dist/learn-system-design /usr/bin/learn-system-design
USER app
CMD [ "learn-system-design", "serve" ]
