FROM golang:1.18.3-bullseye as builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ./app cmd/server-manager/main.go

FROM pulumi/pulumi-go


COPY --from=builder /build/app /app

ENTRYPOINT ["/app"]