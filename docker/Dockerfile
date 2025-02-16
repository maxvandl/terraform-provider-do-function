FROM golang:1.19.3-alpine3.15

ARG BINARY
ENV BINARY=${BINARY}

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./

RUN mkdir klayers
COPY klayers/*.go ./klayers/

RUN go build -o ${BINARY}