ARG GO_VERSION=1.25-alpine
ARG PYTHON_VERSION=3.12
ARG ALPINE_VERSION=3.22

FROM golang:${GO_VERSION} AS pandas-ta-builder

RUN apk add --no-cache build-base git

WORKDIR /src

# pandas-ta runtime 依赖 workspace 本地模块和后端源码，容器内必须保留仓库结构。
COPY go.work ./go.work
COPY go.work.sum ./go.work.sum
COPY proto /src/proto
COPY app/backend /src/app/backend

ENV GOWORK=/src/go.work

WORKDIR /src/app/backend/pandas-ta
RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/pandas-ta-runtime ./cmd/pandas-ta-runtime

FROM python:${PYTHON_VERSION}-slim AS pandas-ta-python-builder

WORKDIR /build
COPY app/backend/pandas-ta/pyproject.toml /build/pyproject.toml
COPY app/backend/pandas-ta/requirements.txt /build/requirements.txt
RUN python -m venv /opt/pandas-ta-venv
ENV PATH="/opt/pandas-ta-venv/bin:${PATH}"
RUN pip install --no-cache-dir --upgrade pip setuptools wheel \
    && pip install --no-cache-dir -r /build/requirements.txt

FROM python:${PYTHON_VERSION}-slim AS pandas-ta-release

RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    netcat-openbsd \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=pandas-ta-builder /out/pandas-ta-runtime ./pandas-ta-runtime
COPY --from=pandas-ta-python-builder /opt/pandas-ta-venv /opt/pandas-ta-venv
COPY app/backend/pandas-ta/scripts /app/app/backend/pandas-ta/scripts

ENV PATH="/opt/pandas-ta-venv/bin:${PATH}"
ENV PANDAS_TA_RUNTIME_PORT=50131
ENV PANDAS_TA_PYTHON_PATH=/opt/pandas-ta-venv/bin/python

EXPOSE 50131

HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
  CMD nc -z 127.0.0.1 50131 || exit 1

CMD ["./pandas-ta-runtime"]
