FROM golang:1.22 as builder
WORKDIR /api
COPY go.mod go.sum Makefile ./
RUN make deps
COPY . .
RUN make build
CMD ["./out/api"]

