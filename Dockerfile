FROM golang:latest AS builder

WORKDIR /app

COPY go.mod .
RUN go mod download

COPY . /app
RUN CGO_ENABLED=0 go build -v ./cmd/bot

FROM scratch

COPY --from=builder /app/bot .
COPY --from=builder /app/config.yaml .
CMD /bot -c config.yaml

