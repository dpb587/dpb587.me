FROM dpb587/gget as gget
ARG hugo_version=1.0.1
RUN gget github.com/google/guetzli@v${hugo_version} guetzli=guetzli_linux_x86-64 --executable

FROM golang AS service
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
ENV CGO_ENABLED=0
RUN go build \
  -o exec \
  .

FROM ubuntu
RUN true \
  && apt-get update \
  && apt-get install -y \
    ca-certificates \
    libpng-dev \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /public
COPY --from=service /workspace/exec /usr/bin/imgrewrite
COPY --from=gget /result/guetzli /usr/bin/guetzli
