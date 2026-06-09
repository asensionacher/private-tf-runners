FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o runner ./cmd/runner

# --- Final Stage ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates git curl unzip bash

# 1. Create a system group and user for non-root execution
RUN addgroup -S runner && adduser -S runner -G runner

# Install Terraform
RUN curl -fsSL https://releases.hashicorp.com/terraform/1.13.0/terraform_1.13.0_linux_amd64.zip -o terraform.zip && \
    unzip terraform.zip && \
    mv terraform /usr/local/bin/ && \
    rm terraform.zip && \
    terraform -version

# Install OpenTofu
RUN curl -L -o tofu.zip "https://github.com/opentofu/opentofu/releases/download/v1.12.1/tofu_1.12.1_linux_amd64.zip" && \
    unzip tofu.zip && \
    mv tofu /usr/local/bin/ && \
    rm tofu.zip && \
    tofu --version

# System-wide certificates (configured as root, which is correct)
COPY ./nginx/certs/* /usr/local/share/ca-certificates/
RUN update-ca-certificates

# 2. Copy configuration files and assign ownership to the new user
COPY --chown=runner:runner ./cmd/runner/rcfiles/* /home/runner/

WORKDIR /app

# 3. Copy the binary and assign ownership to the new user
COPY --chown=runner:runner --from=builder /app/runner .
RUN chmod +x /app/runner

# 4. Switch to the non-privileged user for the rest of the execution
USER runner

ENTRYPOINT ["./runner"]