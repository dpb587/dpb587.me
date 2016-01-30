FROM ubuntu:wily

ENV LC_CTYPE "en_US.UTF-8"
ENV LANG "en_US.UTF-8"
RUN \
  locale-gen "en_US.UTF-8" \
  && apt-get update \
  && apt-get install -y \
    git \
    ruby \
    ruby-dev \
    make \
    gcc \
    nodejs \
    python-pygments \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/ \
  && gem install bundler
