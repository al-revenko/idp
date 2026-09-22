FROM golang:1.27 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/idp/ cmd/idp/
COPY internal/ internal/
COPY proto/gen/ proto/gen/
COPY db/sql/gen/ db/sql/gen/

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/idp ./cmd/idp

FROM alpine:3.22

RUN adduser -D -u 1000 idp

COPY --from=build /out/idp /usr/local/bin/idp

USER idp

ENV HTTP_ADDR=0.0.0.0:50050 \
    GRPC_ADDR=0.0.0.0:50051

EXPOSE 50050 50051

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:50050/health || exit 1

ENTRYPOINT ["/usr/local/bin/idp"]
