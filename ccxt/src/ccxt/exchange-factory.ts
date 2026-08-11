import ccxt from "ccxt";
import type { Exchange, ExchangeOptions } from "ccxt";

import type { ExchangeCredentials } from "../types/proto.js";
import { logger } from "../logger.js";
import {
  EXCHANGE_TYPE_BINANCE,
  EXCHANGE_TYPE_BITGET,
  EXCHANGE_TYPE_BYBIT,
  EXCHANGE_TYPE_OKX,
  SUPPORTED_EXCHANGE_TYPES,
} from "./constants.js";
import { normalizeExchangeType } from "./helpers.js";

export type RuntimeExchange = Exchange;

type ExchangeClass = new (config?: ExchangeOptions) => Exchange;

function applyExchangeProxy(config: ExchangeOptions, proxy: string): void {
  if (!proxy.trim()) {
    return;
  }
  config.httpsProxy = proxy.trim();
}

function exchangeClassOf(exchangeType: string): ExchangeClass {
  switch (exchangeType) {
    case EXCHANGE_TYPE_BINANCE:
      return ccxt.binance;
    case EXCHANGE_TYPE_BYBIT:
      return ccxt.bybit;
    case EXCHANGE_TYPE_OKX:
      return ccxt.okx;
    case EXCHANGE_TYPE_BITGET:
      return ccxt.bitget;
    default:
      throw new Error(`不支持的交易所类型: ${exchangeType}`);
  }
}

function buildOptions(exchangeType: string, credentials: ExchangeCredentials, proxy: string): ExchangeOptions {
  const apiKey = (credentials.api_key ?? "").trim();
  const secret = (credentials.secret_key ?? "").trim();
  if (!apiKey || !secret) {
    throw new Error("交易所 API Key 或 Secret 未配置");
  }

  // 统一在这里构造交易所连接配置，避免不同入口的参数漂移。
  const options: ExchangeOptions = {
    apiKey,
    secret,
    enableRateLimit: true,
    options: { defaultType: "future" },
  };
  applyExchangeProxy(options, proxy);

  if (exchangeType === EXCHANGE_TYPE_OKX && (credentials.passphrase ?? "").trim()) {
    options.password = credentials.passphrase?.trim();
  }
  if (exchangeType === EXCHANGE_TYPE_BYBIT) {
    options.options = {
      defaultType: "future",
      fetchMarkets: { types: ["linear"] },
    };
  }
  return options;
}

function applySandbox(exchange: RuntimeExchange, exchangeType: string, credentials: ExchangeCredentials): void {
  // demo 与 testnet 是两种不同的目标环境，必须显式区分，避免把 testnet 密钥错误路由到 demo 域名。
  if (credentials.is_demo) {
    if (exchangeType === EXCHANGE_TYPE_BINANCE || exchangeType === EXCHANGE_TYPE_BYBIT) {
      const target = exchange as RuntimeExchange & { enableDemoTrading?: (enabled: boolean) => void };
      target.enableDemoTrading?.(true);
      return;
    }
    exchange.setSandboxMode?.(true);
    return;
  }

  // 对支持标准沙盒模式的交易所，testnet 统一走 sandbox 开关。
  if (credentials.testnet) {
    exchange.setSandboxMode?.(true);
  }
}

export async function ensureMarketsLoaded(exchange: RuntimeExchange): Promise<void> {
  let lastError: unknown;
  for (let attempt = 0; attempt < 3; attempt += 1) {
    try {
      const markets = await exchange.loadMarkets();
      if (markets && Object.keys(markets).length > 0) {
        return;
      }
      lastError = new Error("empty market set");
    } catch (error) {
      lastError = error;
      logger.error("加载市场元数据失败，准备重试", { attempt: attempt + 1, error: String(error) });
    }
    if (attempt < 2) {
      await new Promise((resolve) => setTimeout(resolve, 250));
    }
  }
  throw new Error(`加载市场元数据失败: ${String(lastError)}`);
}

export async function createPrivateExchange(
  credentials: ExchangeCredentials | undefined,
  proxy: string,
): Promise<{ exchange: RuntimeExchange; exchangeType: string }> {
  if (!credentials) {
    throw new Error("缺少交易所凭证");
  }
  const exchangeType = normalizeExchangeType(credentials.exchange_type);
  if (!SUPPORTED_EXCHANGE_TYPES.has(exchangeType)) {
    throw new Error(`不支持的交易所类型: ${exchangeType}`);
  }

  const ExchangeClass = exchangeClassOf(exchangeType);
  const exchange = new ExchangeClass(buildOptions(exchangeType, credentials, proxy));
  applySandbox(exchange, exchangeType, credentials);
  await ensureMarketsLoaded(exchange);
  return { exchange, exchangeType };
}

export class PublicExchangeFactory {
  private readonly proxy: string;
  private readonly cache = new Map<string, RuntimeExchange>();

  constructor(proxy: string) {
    this.proxy = proxy.trim();
  }

  async get(exchangeType: string): Promise<RuntimeExchange> {
    const normalized = normalizeExchangeType(exchangeType);
    const cached = this.cache.get(normalized);
    if (cached) {
      return cached;
    }

    const ExchangeClass = exchangeClassOf(normalized);
    const options: ExchangeOptions = {};
    applyExchangeProxy(options, this.proxy);
    const exchange = new ExchangeClass(options);
    await ensureMarketsLoaded(exchange);
    this.cache.set(normalized, exchange);
    return exchange;
  }
}
