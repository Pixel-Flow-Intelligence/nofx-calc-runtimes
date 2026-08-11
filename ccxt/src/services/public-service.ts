import type { Market, OHLCV } from "ccxt";

import { PublicExchangeFactory } from "../ccxt/exchange-factory.js";
import {
  EXCHANGE_TYPE_BINANCE,
} from "../ccxt/constants.js";
import {
  mapKline,
} from "../ccxt/mappers.js";
import {
  normalizeContractSymbol,
  normalizeExchangeType,
  normalizeSpotSymbol,
  timeframeToMillis,
  toNumber,
} from "../ccxt/helpers.js";
import type {
  FetchFundingRateRequest,
  FetchKlinesRequest,
  FetchOpenInterestRequest,
  FetchSymbolTradingRulesRequest,
  GetCurrentPriceRequest,
} from "../types/proto.js";

type SymbolTradingRulesPayload = {
  symbol: string;
  min_order_amount: number;
  max_order_amount: number;
  market_min_order_amount: number;
  amount_step: number;
  min_notional: number;
  price_step: number;
  min_price: number;
  max_price: number;
  max_leverage: number;
};

export class PublicMarketService {
  constructor(private readonly factory: PublicExchangeFactory) {}

  async getCurrentPrice(request: GetCurrentPriceRequest): Promise<{ price: number }> {
    const exchange = await this.factory.get(this.defaultExchange(request.exchange));
    const ticker = await exchange.fetchTicker(normalizeSpotSymbol(request.symbol));
    const price = toNumber(ticker.last) || toNumber(ticker.bid) || toNumber(ticker.ask);
    if (!price) {
      throw new Error("ticker 中无有效价格");
    }
    return { price };
  }

  async fetchKlines(request: FetchKlinesRequest): Promise<{ klines: ReturnType<typeof mapKline>[] }> {
    const exchange = await this.factory.get(request.exchange ?? "");
    const timeframe = (request.timeframe ?? "").trim() || "1m";
    const limit = toNumber(request.limit) || 0;
    const timeframeMs = timeframeToMillis(timeframe);
    const rows: OHLCV[] = await exchange.fetchOHLCV(normalizeSpotSymbol(request.symbol), timeframe, undefined, limit);
    return { klines: rows.map((item) => mapKline(item, timeframeMs)) };
  }

  async fetchOpenInterest(request: FetchOpenInterestRequest): Promise<{ latest: number; average: number }> {
    const exchange = await this.factory.get(request.exchange ?? "");
    const payload = await exchange.fetchOpenInterest(normalizeContractSymbol(request.symbol));
    const latest = toNumber(payload.openInterestAmount) || toNumber(payload.openInterestValue);
    return { latest, average: latest };
  }

  async fetchFundingRate(request: FetchFundingRateRequest): Promise<{ rate: number }> {
    const exchange = await this.factory.get(request.exchange ?? "");
    const payload = await exchange.fetchFundingRate(normalizeContractSymbol(request.symbol));
    return { rate: toNumber(payload.fundingRate) };
  }

  async fetchSymbolTradingRules(request: FetchSymbolTradingRulesRequest): Promise<{ rules: SymbolTradingRulesPayload[] }> {
    const exchange = await this.factory.get(this.defaultExchange(request.exchange));
    const rules: SymbolTradingRulesPayload[] = [];
    for (const symbol of request.symbols ?? []) {
      const market = this.readMarket(exchange, symbol);
      if (!market) {
        continue;
      }
      rules.push(this.mapTradingRules(market, symbol));
    }
    return { rules };
  }

  private defaultExchange(exchange?: string): string {
    return normalizeExchangeType(exchange) || EXCHANGE_TYPE_BINANCE;
  }

  private readMarket(exchange: Awaited<ReturnType<PublicExchangeFactory["get"]>>, symbol?: string): Market | undefined {
    const normalizedContract = normalizeContractSymbol(symbol);
    const normalizedSpot = normalizeSpotSymbol(symbol);
    for (const candidate of [normalizedContract, normalizedSpot]) {
      if (!candidate) {
        continue;
      }
      try {
        return exchange.market(candidate);
      } catch {
        // 不同交易所可能把同一交易对暴露为 spot 或 swap，这里继续尝试下一个标准化形态。
      }
    }
    return undefined;
  }

  private mapTradingRules(market: Market, requestedSymbol?: string): SymbolTradingRulesPayload {
    const amountLimits = market.limits?.amount ?? {};
    const marketLimits = market.limits?.market ?? {};
    const priceLimits = market.limits?.price ?? {};
    const costLimits = market.limits?.cost ?? {};
    const leverageLimits = market.limits?.leverage ?? {};
    return {
      symbol: normalizeSpotSymbol(requestedSymbol || market.symbol),
      min_order_amount: toNumber(amountLimits.min),
      max_order_amount: toNumber(amountLimits.max),
      market_min_order_amount: toNumber(marketLimits.min),
      amount_step: toNumber(market.precision?.amount),
      min_notional: toNumber(costLimits.min),
      price_step: toNumber(market.precision?.price),
      min_price: toNumber(priceLimits.min),
      max_price: toNumber(priceLimits.max),
      max_leverage: toNumber(leverageLimits.max),
    };
  }
}
