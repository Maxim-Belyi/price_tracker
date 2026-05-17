FROM golang:1.26-alpine AS builder
 
WORKDIR /app

COPY go.mod go.sum ./
RUN go mid download

COPY . .

RUN go build -o worker_app ./worker/cmd/

FROM alpine:latest