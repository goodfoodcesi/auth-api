FROM golang:1.23.4-alpine AS builder
WORKDIR /app

RUN apk add --no-cache openssl

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w" -o /authapi main.go


FROM scratch AS build-release-stage

WORKDIR /

COPY --from=builder /authapi /authapi


EXPOSE 8080

ENTRYPOINT ["/authapi"]