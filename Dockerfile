FROM golang:1.23-alpine AS builder
COPY . .
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -installsuffix nocgo -o /toast ./cmd/toast

FROM scratch
COPY --from=builder /toast /toast
ENTRYPOINT ["/toast"]

