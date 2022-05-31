FROM golang:alpine AS builder
RUN apk --update add ca-certificates
RUN mkdir -p /app
WORKDIR /app
ADD go.mod /app
RUN go mod download

ADD . /app
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -a -o /app/nats-tail

#RUN ln -sf /dev/stdout /var/log/nginx/access.log \
#&& ln -sf /dev/stderr /var/log/nginx/error.log
FROM scratch
ARG NATS_TOKEN
ENV NATS_TOKEN=${NATS_TOKEN}
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/nats-tail /nats-tail
ENTRYPOINT ["/nats-tail"]
