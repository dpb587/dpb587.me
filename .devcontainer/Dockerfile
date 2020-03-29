FROM ubuntu:latest

ARG HUGO_VERSION=0.62.1

RUN true \
  && apt update \
  && apt --assume-yes install ca-certificates git wget \
  && wget -qO- https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_${HUGO_VERSION}_linux-64bit.tar.gz \
     | tar -xzf- -C /usr/local/bin hugo
RUN echo "PS1='\$(pwd | sed "s@^/workspaces/@./@" )\$ '" >> ~/.bashrc

EXPOSE 1313
