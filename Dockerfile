FROM golang:1.20.4-alpine3.18 AS builder

ENV GO111MODULE on
RUN unset GOPATH
RUN echo ${GOPATH}

RUN mkdir /build
WORKDIR /build

COPY ./ ./
RUN ls -lh

RUN go mod tidy
RUN go build -o ./gomigrate cmd/main.go

FROM alpine:3.18.0

RUN addgroup -S gomigrate && adduser -S gomigrate -G gomigrate
USER gomigrate
WORKDIR /home/gomigrate

COPY --from=builder /build/gomigrate /home/gomigrate/gomigrate

CMD ["/home/gomigrate/gomigrate"]
