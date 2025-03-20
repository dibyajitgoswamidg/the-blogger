FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o blogger ./cmd/api/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/blogger .
CMD ["./blogger"]