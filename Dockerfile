FROM golang:1.27-alpine AS builder
LABEL authors="alice088 with love for БЮРО 1440"

WORKDIR /src

COPY go.sum ./
COPY go.mod ./

RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /app ./cmd/scpi-sim1440/main.go

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app /app
COPY --from=builder /src/config/devices.yaml /config/devices.yaml

WORKDIR /
USER nonroot:nonroot
ENTRYPOINT ["/app"]