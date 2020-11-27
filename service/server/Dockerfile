FROM golang AS server
ARG BUILD_COMMIT
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
ENV CGO_ENABLED=0
RUN go build \
  -o exec \
  -ldflags " \
    -s -w \
    -X main.appCommit=$BUILD_COMMIT \
    -X main.appBuilt=$( date -u +%Y-%m-%dT%H:%M:%SZ ) \
  " \
  .

FROM alpine
RUN apk add --no-cache ca-certificates
WORKDIR /dpb587.me
COPY --from=server /workspace/exec bin/server
ENV PORT 8080
EXPOSE 8080
ENTRYPOINT [ "/dpb587.me/bin/server" ]
