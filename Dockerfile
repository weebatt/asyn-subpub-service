FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /server cmd/server/main.go

FROM alpine:latest
COPY --from=builder /server /server
COPY config/config.yaml /config.yaml
CMD ["/server", "-config", "/config.yaml"]