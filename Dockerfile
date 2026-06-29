FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o webhook .

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/webhook .
USER 65532:65532
EXPOSE 443
ENTRYPOINT ["/webhook"]
