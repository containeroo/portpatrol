FROM golang:1.23-alpine AS builder
COPY . .
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -installsuffix nocgo -o /thor ./cmd/thor

FROM scratch
COPY --from=builder /thor /thor
ENTRYPOINT ["/thor"]

