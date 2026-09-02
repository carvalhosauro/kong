# Demo image: Debian with real cron jobs + kong binary for README screenshots.
FROM golang:1.22-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=demo" -o /kong ./cmd/kong

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    cron \
    procps \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Users with crontabs
RUN useradd -m deploy && useradd -m alice

COPY docker/demo/etc/crontab /etc/crontab
COPY docker/demo/etc/cron.d/ /etc/cron.d/
COPY docker/demo/var/spool/cron/crontabs/ /var/spool/cron/crontabs/
COPY docker/demo/etc/cron.daily/ /etc/cron.daily/
COPY docker/demo/etc/cron.hourly/ /etc/cron.hourly/
COPY docker/demo/etc/anacrontab /etc/anacrontab

# Debian cron expects spool files owned by root:crontab
RUN chown -R root:crontab /var/spool/cron/crontabs \
    && chmod 600 /var/spool/cron/crontabs/* \
    && chmod 644 /etc/cron.d/* \
    && chmod +x /etc/cron.daily/* /etc/cron.hourly/*

COPY --from=builder /kong /usr/local/bin/kong

# Fixed timezone for reproducible NEXT timestamps in screenshots
ENV TZ=America/Sao_Paulo
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /root
CMD ["kong", "list"]
