FROM golang:1.24-bookworm AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY web ./web

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/server ./cmd/server

FROM gcr.io/distroless/static-debian12
WORKDIR /app

COPY --from=builder /bin/server /app/server
COPY --from=builder /app/web /app/web

EXPOSE 8080

ENV PORT=8080
ENV JWT_SECRET=development-secret
ENV JWT_ISSUER=interviewsw

ENTRYPOINT ["/app/server"]
