FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/klevliga ./cmd/klevliga

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
COPY --from=build /out/klevliga /bin/app
COPY config.docker.yaml /app/config.yaml
EXPOSE 8080
ENTRYPOINT ["/bin/app"]
CMD ["-c", "/app/config.yaml"]
