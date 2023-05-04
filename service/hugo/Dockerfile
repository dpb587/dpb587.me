FROM dpb587/gget as gget
ARG hugo_version=0.81.0
RUN gget github.com/gohugoio/hugo@v${hugo_version} 'hugo_extended_*_Linux-64bit.tar.gz' --stdout | tar -xzf- hugo

FROM node:20-bullseye AS hugo
RUN true \
  && export DEBIAN_FRONTEND=noninteractive \
  && apt-get update \
  && apt-get install -y \
    ca-certificates \
    git \
  && rm -rf /var/lib/apt/lists/*
COPY --from=gget /result/hugo /usr/bin/hugo
WORKDIR /hugo
EXPOSE 1313
ADD hugo-with-env /usr/bin/hugo-with-env
ENTRYPOINT [ "/usr/bin/hugo-with-env" ]
