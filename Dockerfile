ARG GO_VERSION=latest
ARG ALPINE_VERSION=latest

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG VERSION=dev

RUN CGO_ENABLED=0 \
  GOOS=linux \
  go build \
  -ldflags "-s -w -X main.Version=${VERSION}" \
  -o local-clipboard .

FROM alpine:${ALPINE_VERSION}

WORKDIR /app

COPY --from=builder /app/local-clipboard .

EXPOSE 8080

# Set this to the host IP/hostname that clients should use to reach the server.
# In Docker the container IP is usually not reachable from the host LAN, so pass
# the host's address with: -e LOCAL_CLIPBOARD_HOST=192.168.1.100
ENV LOCAL_CLIPBOARD_HOST=""

ENTRYPOINT ["./local-clipboard"]
