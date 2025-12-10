FROM golang:1.23-alpine as build
RUN apk add --no-cache tzdata
WORKDIR /app
COPY go.mod go.sum vendor main.go ./
RUN go build -o app .


FROM alpine:latest
RUN apk add --no-cache tzdata
RUN apk --no-cache add ca-certificates
COPY --from=build /app/app /app
CMD ["/app"]
