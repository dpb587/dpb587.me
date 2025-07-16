FROM debian:bookworm-slim AS base
ARG TARGETARCH
RUN apt update \
    && apt install -y \
        brotli \
        ca-certificates \
        git \
        wget \
        zstd \
    && rm -rf /var/lib/apt/lists/*
RUN wget -qO- https://github.com/gohugoio/hugo/releases/download/v0.147.9/hugo_0.147.9_linux-${TARGETARCH}.tar.gz \
    | tar -xzf- -C /usr/local/bin hugo
RUN wget -qO- https://go.dev/dl/go1.24.4.linux-${TARGETARCH}.tar.gz \
    | tar -C /usr/local -xzf-
ENV PATH="/usr/local/go/bin:${PATH}"

FROM base AS hugo-build
ADD . /workspaces/main
WORKDIR /workspaces/main
RUN cd hugo \
    && hugo
RUN ./scripts/package-public.sh hugo/public
RUN ( echo -n 'time ' ; date -u +%Y-%m-%dT%H:%M:%SZ ; echo -n 'commit ' ; git rev-parse HEAD ; hugo version ) > hugo/public/internal.txt
RUN cd tools/publish \
    && go build -o ./server ./cmd/server

FROM debian:bookworm-slim
RUN apt update \
    && apt install -y \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=hugo-build /workspaces/main/tools/publish/server /deploy/bin/server
COPY --from=hugo-build /workspaces/main/hugo/public /deploy/docroot/
ADD data/redirects /deploy/etc/redirects
CMD ["/deploy/bin/server", "/deploy/docroot", "/deploy/etc/redirects"]
EXPOSE 8080
