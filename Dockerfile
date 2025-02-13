FROM golang:1.24-alpine AS builder
COPY . .
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -installsuffix nocgo -o /portpatrol ./cmd/portpatrol

FROM scratch
COPY --from=builder /portpatrol /portpatrol
ENTRYPOINT ["/portpatrol"]

