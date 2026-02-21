FROM mcr.microsoft.com/devcontainers/go:1-trixie

RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    libopus-dev \
    libopusfile-dev \
    && rm -rf /var/lib/apt/lists/*
