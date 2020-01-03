FROM golang:1.13 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY k8s/reputator .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin .
COPY --from=builder /app/resources ./resources/

ENV PORT 8080
ENV DEBUG 1

CMD ["./reputator"]
