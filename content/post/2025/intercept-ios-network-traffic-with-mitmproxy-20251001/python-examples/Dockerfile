FROM python:3.11-slim
WORKDIR /app
RUN pip install --no-cache-dir mitmproxy
COPY addon.py .
ENV PYTHONUNBUFFERED=1
