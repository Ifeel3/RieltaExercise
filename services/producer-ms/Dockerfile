FROM golang:alpine
RUN mkdir /build
RUN mkdir -p /app/logs
WORKDIR /build
COPY . /build/
RUN go build -o /app/producer-ms ./cmd/producer-ms/
WORKDIR /app
RUN rm -rf /build
CMD ./producer-ms