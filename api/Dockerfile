FROM golang:1.23.6-alpine3.21 AS build
WORKDIR /app

COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o myapp .

FROM alpine:3.21
WORKDIR /app

COPY --from=build /app/myapp .
ENTRYPOINT ["/app/myapp"]
