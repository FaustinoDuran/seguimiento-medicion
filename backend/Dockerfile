FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -o /mcp ./cmd/mcp

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /mcp /mcp
USER 65532:65532
ENV MCP_TRANSPORT=http PORT=8080
EXPOSE 8080
ENTRYPOINT ["/mcp"]
