FROM golang:alpine
RUN mkdir /build
RUN mkdir -p /app/logs
WORKDIR /build
COPY . /build/
RUN go build -o /app/consumer-ms ./cmd/consumer-ms
WORKDIR /app
RUN rm -rf /build
CMD ./consumer-ms