FROM docker.io/golang:1.18 AS build

ENV CGO_ENABLED 0

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN make build

FROM scratch

COPY --from=build /app/main /main

EXPOSE 8080

CMD ["/main"]
