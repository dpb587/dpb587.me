FROM golang:1.12 as build-go
WORKDIR /go/src/github.com/dpb587/dpb587.me
COPY server ./server
RUN mkdir /build
ENV CGO_ENABLED=0
RUN go build -o /build/exec ./server/cmd/web

FROM scratch
COPY content /app/content
COPY theme /app/theme
COPY --from=build-go /build/exec /app/exec
CMD ["/app/exec"]
