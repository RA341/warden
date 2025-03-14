FROM golang:1.24-alpine AS builder

# Build arguments
ARG VERSION=dev
ARG COMMIT_INFO=unknown
ARG BUILD_DATE=unknown
ARG BRANCH=unknown

# for sqlite
RUN apk update && apk add --no-cache gcc musl-dev
ENV CGO_ENABLED=1

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# arg substitution, do not put it higher than this for caching
# https://stackoverflow.com/questions/44438637/arg-substitution-in-run-command-not-working-for-dockerfile
ENV VERSION=${VERSION}
ENV COMMIT_INFO=${COMMIT_INFO}
ENV BUILD_DATE=${BUILD_DATE}
ENV BRANCH=${BRANCH}

# build optimized binary without debugging symbols
RUN SOURCE_HASH=$(find . -type f -name "*.go" -print0 | sort -z | xargs -0 cat | sha256sum | cut -d ' ' -f1) && \
    go build -ldflags "-s -w \
      -X github.com/RA341/warden/Version=${VERSION} \
      -X github.com/RA341/warden/CommitInfo=${COMMIT_INFO} \
      -X github.com/RA341/warden/BuildDate=${BUILD_DATE} \
      -X github.com/RA341/warden/Branch=${BRANCH} \
      -X github.com/RA341/warden/SourceHash=${SOURCE_HASH}" \
    -o warden

FROM alpine:latest

WORKDIR /app/

COPY --from=builder /app/warden .

EXPOSE 8080

CMD ["./warden"]
