FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /nanogpt-proxy .

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /nanogpt-proxy /nanogpt-proxy
ENV PORT=8080 \
    TARGET_URL=https://nano-gpt.com
EXPOSE 8080
USER 65534:65534
ENTRYPOINT ["/nanogpt-proxy"]
