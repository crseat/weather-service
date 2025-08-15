# ---- Build stage ----
FROM golang:1.22 AS build
WORKDIR /src
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/weatherd ./cmd/weatherd

# ---- Runtime stage ----
FROM gcr.io/distroless/base-debian12
USER nonroot:nonroot
COPY --from=build /out/weatherd /usr/local/bin/weatherd
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/weatherd"]
