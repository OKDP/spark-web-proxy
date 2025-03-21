ARG GO_VERSION=1.23

FROM golang:${GO_VERSION} AS go-build

ARG GIT_COMMIT="_unset_"
ARG LDFLAGS="-X localbuild=true"
ARG TARGETOS="linux"
ARG TARGETARCH

WORKDIR /workspace/spark-web-proxy

COPY Makefile Makefile
COPY go.* ./
COPY *.go ./
COPY internal/ internal/
COPY cmd/ cmd/

RUN go mod tidy \
    && go mod download
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    LDFLAGS=${LDFLAGS##-X localbuild=true} GIT_COMMIT=$GIT_COMMIT \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o spark-web-proxy main.go

FROM alpine:3.20.3

RUN apk --no-cache add ca-certificates && update-ca-certificates

COPY --from=go-build /workspace/spark-web-proxy /usr/local/bin/

USER 65534:65534

EXPOSE 8090

ENTRYPOINT ["spark-web-proxy"]

