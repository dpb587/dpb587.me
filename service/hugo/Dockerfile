FROM dpb587/gget as gget
ARG hugo_version=0.62.1
RUN gget github.com/gohugoio/hugo@v${hugo_version} 'hugo_extended_*_Linux-64bit.tar.gz' --stdout | tar -xzf- hugo

FROM ubuntu AS hugo
RUN true \
  && apt-get update \
  && apt-get install -y \
    ca-certificates \
    git \
  && rm -rf /var/lib/apt/lists/*
COPY --from=gget /result/hugo /bin/hugo
WORKDIR /hugo
EXPOSE 1313
ENTRYPOINT [ "hugo" ]
