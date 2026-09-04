FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN go build -ldflags="-s -w" -o /json-doctor .

FROM alpine:3.20
COPY --from=build /json-doctor /usr/local/bin/json-doctor
ENTRYPOINT ["json-doctor"]
