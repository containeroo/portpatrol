FROM golang:1.23-alpine AS builder
COPY . .
RUN CGO_ENABLED=0 GO111MODULE=on go build -ldflags="-w -s" -o /toast ./cmd/toast

FROM scratch
COPY --from=builder /toast /toast
ENTRYPOINT ["/toast"]

