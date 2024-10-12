FROM golang:alpine AS builder
WORKDIR /app

ADD go.mod .
ADD go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build \
    -tags release \
    -o ./bin/api ./cmd/api

FROM scratch AS api
EXPOSE 3000
COPY --from=builder /app/bin/api /
CMD ["/api"]
