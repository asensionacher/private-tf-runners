FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o runner ./cmd/runner

FROM alpine:3.19

RUN apk add --no-cache ca-certificates git curl unzip bash

RUN curl -fsSL https://releases.hashicorp.com/terraform/1.13.0/terraform_1.13.0_linux_amd64.zip -o terraform.zip && \
    unzip terraform.zip && \
    mv terraform /usr/local/bin/ && \
    rm terraform.zip && \
    terraform -version

RUN curl -L -o tofu.zip "https://github.com/opentofu/opentofu/releases/download/v1.12.1/tofu_1.12.1_linux_amd64.zip" && \
    unzip tofu.zip && \
    mv tofu /usr/local/bin/ && \
    rm tofu.zip && \
    tofu --version

COPY ../../nginx/certs/* /usr/local/share/ca-certificates/
RUN update-ca-certificates

WORKDIR /app

COPY --from=builder /app/runner .

RUN chmod +x /app/runner

ENTRYPOINT ["./runner"]