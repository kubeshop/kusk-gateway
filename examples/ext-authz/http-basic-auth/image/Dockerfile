FROM golang:1.18-alpine AS build

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /app

COPY . .
RUN go mod download
RUN go mod verify

RUN go build -tags 'osusergo netgo' -ldflags '-extldflags "-static" -s -w' -a -v -o ./main ./main.go

FROM scratch

COPY --from=build /app/main /usr/bin/main

EXPOSE 8080

CMD ["/usr/bin/main"]
