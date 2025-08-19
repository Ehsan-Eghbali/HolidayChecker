FROM golang:1.25 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go test ./... -race -v
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/holidaychecker ./cmd/holidaychecker

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/holidaychecker /holidaychecker
ENTRYPOINT ["/holidaychecker"]