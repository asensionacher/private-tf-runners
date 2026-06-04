# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git gcc musl-dev

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o /server ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates sqlite-libs git

WORKDIR /app

COPY --from=builder /server .

RUN mkdir -p /data

EXPOSE 8080

ENV DATABASE_PATH=/data/runners.db
ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8080
ENV ENV=production

ENTRYPOINT ["./server"]
