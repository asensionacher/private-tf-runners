FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o runner ./cmd/runner

FROM alpine:3.19

RUN apk add --no-cache ca-certificates git curl unzip bash

RUN curl -fsSL https://releases.hashicorp.com/terraform/1.6.6/terraform_1.6.6_linux_amd64.zip -o terraform.zip && \
    unzip terraform.zip && \
    mv terraform /usr/local/bin/ && \
    rm terraform.zip && \
    terraform -version

RUN curl -L -o tofu.zip "https://github.com/opentofu/opentofu/releases/download/v1.12.1/tofu_1.12.1_linux_amd64.zip" && \
    unzip tofu.zip && \
    mv tofu /usr/local/bin/ && \
    rm tofu.zip && \
    tofu --version

COPY ../../nginx/certs/api.runners.lan.crt /usr/local/share/ca-certificates/api.runners.lan.crt
COPY ../../nginx/certs/ca.crt /usr/local/share/ca-certificates/ca.crt
COPY ../../nginx/certs/server.crt /usr/local/share/ca-certificates/server.crt
COPY ./miniblue.crt /usr/local/share/ca-certificates/miniblue.crt
RUN update-ca-certificates

WORKDIR /app

COPY --from=builder /app/runner .

RUN chmod +x /app/runner

ENTRYPOINT ["./runner"]