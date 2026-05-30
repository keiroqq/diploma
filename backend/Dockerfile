FROM golang:1.26.2-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download
RUN CGO_ENABLED=0 go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
  -trimpath \
  -ldflags="-s -w" \
  -o /out/content-digest \
  ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
  && addgroup -S app \
  && adduser -S -G app app

WORKDIR /app

COPY --from=build /out/content-digest /app/content-digest
COPY --from=build /go/bin/migrate /app/migrate
COPY --from=build /src/migrations /app/migrations

USER app

EXPOSE 8080

ENTRYPOINT ["/app/content-digest"]
