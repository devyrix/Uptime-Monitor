# ===== Tahap 1: build =====
# Image besar dengan toolchain Go. Hanya dipakai saat build.
FROM golang:1.27-alpine AS build

WORKDIR /app

# Salin definisi dependensi dulu, lalu download terpisah.
# Trik cache: layer ini tidak di-rebuild kalau go.mod/go.sum tidak berubah.
COPY go.mod go.sum ./
RUN go mod download

# Baru salin sisa kode dan build.
# CGO_ENABLED=0 → binary statis murni, jalan di image final tanpa libc.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /uptime-monitor .

# ===== Tahap 2: image final =====
# Image kecil HANYA berisi binary + aset. Toolchain Go tidak ikut.
FROM alpine:3.20

COPY --from=build /uptime-monitor /usr/local/bin/uptime-monitor
COPY templates /app/templates
COPY targets.yaml /app/targets.yaml

WORKDIR /app
EXPOSE 3000

# exec-form CMD supaya SIGTERM dari `docker stop` diterima langsung oleh binary
CMD ["uptime-monitor"]
