# hugo

FROM dpb587/gget as hugo-deps
ARG hugo_version=0.81.0
RUN gget github.com/gohugoio/hugo@v${hugo_version} 'hugo_extended_*_Linux-64bit.tar.gz' --stdout | tar -xzf- hugo

FROM node:16-stretch AS hugo
RUN true \
  && apt-get update \
  && apt-get install -y \
    ca-certificates \
    git \
  && rm -rf /var/lib/apt/lists/*
COPY --from=hugo-deps /result/hugo /usr/bin/hugo
WORKDIR /result
ARG GEOAPIFY_TOKEN
ARG GITHUB_TOKEN
ARG HUGO_PARAMS_COMMENTSREMARKURL
ADD .git .git
ADD appendix appendix
ADD content content
ADD static static
ADD themes themes
ADD config.toml config.toml
ADD service/hugo/hugo-with-env /usr/bin/
RUN /usr/bin/hugo-with-env
RUN ( hugo version ; git rev-parse HEAD ; date -u +%Y-%m-%dT%H:%M:%SZ ) > public/internal/hugo

# imgrewrite

FROM dpb587/gget as imgrewrite-deps
ARG guetzli_version=1.0.1
RUN gget github.com/google/guetzli@v${guetzli_version} guetzli=guetzli_linux_x86-64 --executable

FROM golang AS imgrewrite-build
WORKDIR /workspace
COPY service/imgrewrite/go.mod service/imgrewrite/go.sum ./
RUN go mod download
COPY service/imgrewrite/ ./
ENV CGO_ENABLED=0
RUN go build \
  -o imgrewrite \
  .

FROM h2non/imaginary as imgrewrite
ARG AWS_ACCESS_KEY_ID
ARG AWS_SECRET_ACCESS_KEY
COPY --from=imgrewrite-build /workspace/imgrewrite /usr/local/bin/imgrewrite
COPY --from=imgrewrite-deps /result/guetzli /usr/local/bin/guetzli
COPY --from=hugo /result/public /result
USER root
RUN /usr/local/bin/imgrewrite /result

# server

FROM golang AS server-build
ARG BUILD_COMMIT
WORKDIR /workspace
COPY service/server/go.mod service/server/go.sum ./
RUN go mod download
COPY service/server/ ./
COPY .git ./.git
ENV CGO_ENABLED=0
RUN go build \
  -o exec \
  -ldflags " \
    -s -w \
    -X main.appCommit=$( git rev-parse HEAD | cut -c-10 ) \
    -X main.appBuilt=$( date -u +%Y-%m-%dT%H:%M:%SZ ) \
  " \
  .

FROM alpine AS server
RUN apk add --no-cache ca-certificates
WORKDIR /dpb587.me
COPY --from=server-build /workspace/exec bin/server
ENV PORT 8080
EXPOSE 8080
ENTRYPOINT [ "./bin/server", "docroot" ]

# pdf

FROM dpb587/gget AS pdf-deps
ARG wkhtmltopdf_version=0.12.6-1
RUN gget github.com/wkhtmltopdf/packaging@${wkhtmltopdf_version} wkhtmltox.deb='wkhtmltox_*.focal_amd64.deb'

FROM ubuntu:focal AS pdf
RUN true \
  && apt-get update \
  && apt-get install -y \
    ca-certificates \
    fontconfig \
    libfreetype6 \
    libjpeg-turbo8 \
    libpng16-16 \
    libssl1.1 \
    libx11-6 \
    libxcb1 \
    libxext6 \
    libxrender1 \
    xfonts-75dpi \
    xfonts-base \
  && rm -rf /var/lib/apt/lists/*
COPY --from=pdf-deps /result/wkhtmltox.deb /tmp
RUN dpkg -i /tmp/wkhtmltox.deb && rm /tmp/wkhtmltox.deb
COPY --from=imgrewrite /result /result
COPY --from=server /dpb587.me/bin/server /usr/local/bin/server
ADD scripts/dump-pdf.sh /tmp/dump-pdf.sh
WORKDIR /result
RUN /tmp/dump-pdf.sh

# final

FROM server
COPY --from=pdf /result /dpb587.me/docroot
