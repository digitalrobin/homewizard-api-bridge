FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /out/homewizard-api-bridge .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/homewizard-api-bridge /usr/local/bin/homewizard-api-bridge

ENV BIND_ADDR=:8080
ENV DATA_DIR=/data
EXPOSE 8080

VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/homewizard-api-bridge"]
