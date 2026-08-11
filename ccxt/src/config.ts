import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const CURRENT_DIR = path.dirname(fileURLToPath(import.meta.url));

export type RuntimeConfig = {
  grpcListen: string;
  grpcHost: string;
  grpcPort: number;
  ccxtProxy: string;
  protoPath: string;
  healthProtoPath: string;
};

function resolveRootProtoPath(...segments: string[]): string {
  const candidates = [
    path.resolve(CURRENT_DIR, "../proto", ...segments),
    path.resolve(CURRENT_DIR, "../../../proto", ...segments),
    path.resolve(CURRENT_DIR, "../../../../proto", ...segments),
  ];
  for (const candidate of candidates) {
    if (fs.existsSync(candidate)) {
      return candidate;
    }
  }
  return candidates[0];
}

function resolveBundledProtoPath(...segments: string[]): string {
  const candidates = [
    path.resolve(CURRENT_DIR, "../proto", ...segments),
    path.resolve(CURRENT_DIR, "../../../proto", ...segments),
  ];
  for (const candidate of candidates) {
    if (fs.existsSync(candidate)) {
      return candidate;
    }
  }
  return candidates[0];
}

function parseListenAddress(raw: string): { host: string; port: number } {
  const value = raw.trim() || "127.0.0.1:50061";
  const matched = value.match(/^(.*):(\d+)$/);
  if (!matched) {
    throw new Error(`无效的 trading-runtime ccxt 适配器地址: ${value}`);
  }
  return {
    host: matched[1] || "127.0.0.1",
    port: Number(matched[2]),
  };
}

export function loadConfig(): RuntimeConfig {
  // 兼容历史 compose 字段 CCXT_GRPC_ADDR，避免容器只监听 127.0.0.1 导致跨服务不可达。
  const grpcListen = (
    process.env.CCXT_RUNTIME_GRPC_ADDR ??
    process.env.CCXT_GRPC_ADDR ??
    "127.0.0.1:50061"
  ).trim();
  const parsed = parseListenAddress(grpcListen);
  return {
    grpcListen,
    grpcHost: parsed.host,
    grpcPort: parsed.port,
    ccxtProxy: (process.env.CCXT_PROXY ?? "").trim(),
    // 优先使用随 ccxt-runtime 镜像打包的 proto，避免根目录 proto 被 .dockerignore 裁剪后启动失败。
    // 本地源码运行仍兼容仓库根 proto 目录，保证开发态与 Docker 态都能找到同一份契约。
    protoPath: resolveRootProtoPath("nofx/ccxtruntime/v1/ccxt_runtime.proto"),
    healthProtoPath: resolveBundledProtoPath("grpc/health/v1/health.proto"),
  };
}
