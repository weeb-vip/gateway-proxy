FROM golang:1.18 as builder
WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/cli/

FROM gcr.io/distroless/static-debian10
WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /app/main .
ARG VERSION
ENV APP__VERSION="${VERSION}"
USER nonroot
CMD ["./main", "server", "start"]
