ARG GO_VERSION=1.25-alpine
ARG PYTHON_VERSION=3.12
ARG ALPINE_VERSION=3.22

FROM golang:${GO_VERSION} AS talib-builder

RUN apk add --no-cache build-base git

WORKDIR /src

# ta-lib runtime 依赖 workspace 本地模块与完整后端目录，必须保留相对路径结构。
COPY go.work ./go.work
COPY go.work.sum ./go.work.sum
COPY proto /src/proto
COPY app/backend /src/app/backend

ENV GOWORK=/src/go.work

WORKDIR /src/app/backend/ta-lib
RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/ta-lib-runtime ./talib/cmd/ta-lib-runtime

FROM python:${PYTHON_VERSION}-slim AS talib-python-builder

ENV TA_PREFIX=/opt/ta-lib-core
ENV TA_LIBRARY_PATH=${TA_PREFIX}/lib
ENV TA_INCLUDE_PATH=${TA_PREFIX}/include

RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build/ta-lib-core
RUN curl -fsSL https://prdownloads.sourceforge.net/ta-lib/ta-lib-0.4.0-src.tar.gz \
    | tar xz --strip-components=1 \
    && ./configure --prefix="${TA_PREFIX}" \
    && make \
    && make install

WORKDIR /build/ta-lib-python
RUN python -m venv /opt/ta-lib-venv
ENV PATH="/opt/ta-lib-venv/bin:${PATH}"
# 这里必须匹配已编译的 TA-Lib C 0.4.x；0.6.x 分支要求底层 C 库升级到 0.6.x。
RUN pip install --no-cache-dir --upgrade pip setuptools wheel \
    && pip install --no-cache-dir numpy Cython TA-Lib==0.5.5 \
    && python -c "import numpy, talib; close = numpy.random.random(100); talib.SMA(close)"

FROM python:${PYTHON_VERSION}-slim AS ta-lib-release

RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    netcat-openbsd \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=talib-builder /out/ta-lib-runtime ./ta-lib-runtime
COPY --from=talib-python-builder /opt/ta-lib-core /opt/ta-lib-core
COPY --from=talib-python-builder /opt/ta-lib-venv /opt/ta-lib-venv
COPY app/backend/ta-lib/scripts /app/app/backend/ta-lib/scripts

ENV PATH="/opt/ta-lib-venv/bin:${PATH}"
ENV TA_LIBRARY_PATH=/opt/ta-lib-core/lib
ENV TA_INCLUDE_PATH=/opt/ta-lib-core/include
ENV LD_LIBRARY_PATH=/opt/ta-lib-core/lib
ENV TA_LIB_RUNTIME_PORT=50121
ENV TA_LIB_PYTHON_PATH=/opt/ta-lib-venv/bin/python
ENV TA_LIB_BRIDGE_SCRIPT_PATH=/app/app/backend/ta-lib/scripts/talib_bridge.py

EXPOSE 50121

HEALTHCHECK --interval=30s --timeout=10s --start-period=25s --retries=3 \
  CMD nc -z 127.0.0.1 50121 || exit 1

CMD ["./ta-lib-runtime"]
