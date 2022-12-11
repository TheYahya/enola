FROM golang:1.19-alpine AS build_base
WORKDIR /tmp/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go test -v
WORKDIR /tmp/app/cmd/enola
RUN go build -o ./enola .

FROM alpine:3.17 
COPY --from=build_base /tmp/app/cmd/enola/enola /app/enola
ENTRYPOINT ["/app/enola"]
