FROM golang:1.17-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build -o /img-host-go

FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /app/web /web

COPY --from=build /img-host-go /img-host-go

EXPOSE 5000

USER nonroot:nonroot

CMD ["./img-host-go"]