FROM golang:1.22 as builder
WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -pgo=auto -ldflags="-w -s" -o /go/bin/app ./cmd/api/**.go

FROM gcr.io/distroless/static-debian11
COPY --from=builder /go/bin/app /
CMD ["/app"]
