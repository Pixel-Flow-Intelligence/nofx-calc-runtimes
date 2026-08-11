ARG GO_VERSION=1.25-alpine
ARG PYTHON_VERSION=3.12
ARG ALPINE_VERSION=3.22

FROM golang:${GO_VERSION} AS trendln-builder

RUN apk add --no-cache build-base git

WORKDIR /src

# trendln runtime 依赖 workspace 本地模块和后端源码，镜像内必须保留完整相对路径。
COPY go.work ./go.work
COPY go.work.sum ./go.work.sum
COPY proto /src/proto
COPY app/backend /src/app/backend

ENV GOWORK=/src/go.work

WORKDIR /src/app/backend/trendln
RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/trendln-runtime ./cmd/trendln-runtime

FROM python:${PYTHON_VERSION}-slim AS trendln-python-builder

RUN python -m venv /opt/trendln-venv
ENV PATH="/opt/trendln-venv/bin:${PATH}"
RUN pip install --no-cache-dir --upgrade pip setuptools wheel \
    && pip install --no-cache-dir 'pandas>=2.2,<3.0' trendln

FROM python:${PYTHON_VERSION}-slim AS trendln-release

RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    netcat-openbsd \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=trendln-builder /out/trendln-runtime ./trendln-runtime
COPY --from=trendln-python-builder /opt/trendln-venv /opt/trendln-venv
COPY app/backend/trendln/scripts /app/app/backend/trendln/scripts

ENV PATH="/opt/trendln-venv/bin:${PATH}"
ENV TRENDLN_RUNTIME_PORT=50141
ENV TRENDLN_PYTHON_PATH=/opt/trendln-venv/bin/python

EXPOSE 50141

HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
  CMD nc -z 127.0.0.1 50141 || exit 1

CMD ["./trendln-runtime"]
