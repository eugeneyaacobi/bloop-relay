FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/bloop-relay ./cmd/bloop-relay

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/bloop-relay /bloop-relay
ENTRYPOINT ["/bloop-relay"]
