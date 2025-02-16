# Production Build Stage
FROM golang:1.22.5 AS build
WORKDIR /app

COPY . ./
RUN go mod download && go mod verify

RUN useradd -u 1001 appuser
RUN go build -ldflags="-linkmode external -extldflags -static" -o ./bin/zmachine


# Production Release Stage
FROM scratch
WORKDIR /app

ENV GIN_MODE=release

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /app/bin/zmachine ./zmachine

USER appuser

ENTRYPOINT ["./zmachine"]