import ccxt from "ccxt";
import type { Position } from "ccxt";

import type {
  CancelAllOrdersRequest,
  CancelOrderRequest,
  ExchangeCredentials,
  FetchBalanceRequest,
  FetchClosedOrdersRequest,
  FetchMyTradesRequest,
  FetchOpenOrdersRequest,
  FetchOrderRequest,
  FetchPositionRequest,
  FetchPositionsHistoryRequest,
  FetchPositionsRequest,
  FetchTickerRequest,
  PlaceLimitOrderRequest,
  PlaceMarketOrderRequest,
  SetLeverageRequest,
  SetMarginModeRequest,
} from "../types/proto.js";
import { createPrivateExchange, type RuntimeExchange } from "../ccxt/exchange-factory.js";
import { EXCHANGE_TYPE_BYBIT } from "../ccxt/constants.js";
import {
  normalizeContractSymbol,
  parseParamsJson,
  toNumber,
} from "../ccxt/helpers.js";
import {
  mapOrder,
  mapPosition,
  mapTicker,
  mapTrade,
} from "../ccxt/mappers.js";

type ExchangeContext = {
  exchange: RuntimeExchange;
  exchangeType: string;
};

export class PrivateExchangeService {
  constructor(private readonly proxy: string) {}

  async fetchBalance(request: FetchBalanceRequest): Promise<{ available_balance: number; margin_used: number; wallet_balance: number }> {
    const { exchange } = await this.open(request.credentials);
    const balance = await exchange.fetchBalance();
    return {
      available_balance: this.readBalanceValue(balance.free, "USDT"),
      margin_used: this.readBalanceValue(balance.used, "USDT"),
      wallet_balance: this.readBalanceValue(balance.total, "USDT"),
    };
  }

  async fetchPositions(request: FetchPositionsRequest): Promise<{ positions: ReturnType<typeof mapPosition>[] }> {
    const context = await this.open(request.credentials);
    const symbols = (request.symbols ?? []).map((item) => normalizeContractSymbol(item));
    const params = this.positionParams(context.exchangeType);
    const rows = await context.exchange.fetchPositions(symbols.length > 0 ? symbols : undefined, undefined, params);
    return { positions: rows.map(mapPosition) };
  }

  async fetchPosition(request: FetchPositionRequest): Promise<{ position: ReturnType<typeof mapPosition> }> {
    const { exchange } = await this.open(request.credentials);
    const position = await exchange.fetchPosition(normalizeContractSymbol(request.symbol));
    return { position: mapPosition(position) };
  }

  async fetchTicker(request: FetchTickerRequest): Promise<{ ticker: ReturnType<typeof mapTicker> }> {
    const { exchange } = await this.open(request.credentials);
    const ticker = await exchange.fetchTicker(normalizeContractSymbol(request.symbol));
    return { ticker: mapTicker(ticker) };
  }

  async placeMarketOrder(request: PlaceMarketOrderRequest): Promise<{ order: ReturnType<typeof mapOrder> }> {
    const { exchange } = await this.open(request.credentials);
    const params = parseParamsJson(request.params_json);
    const order = await exchange.createOrder(
      normalizeContractSymbol(request.symbol),
      "market",
      request.side ?? "",
      toNumber(request.amount),
      undefined,
      params,
    );
    return { order: mapOrder(order) };
  }

  async placeLimitOrder(request: PlaceLimitOrderRequest): Promise<{ order: ReturnType<typeof mapOrder> }> {
    const { exchange } = await this.open(request.credentials);
    const params = parseParamsJson(request.params_json);
    const order = await exchange.createOrder(
      normalizeContractSymbol(request.symbol),
      "limit",
      request.side ?? "",
      toNumber(request.amount),
      toNumber(request.price),
      params,
    );
    return { order: mapOrder(order) };
  }

  async cancelOrder(request: CancelOrderRequest): Promise<{ order: ReturnType<typeof mapOrder> }> {
    const { exchange } = await this.open(request.credentials);
    const symbol = (request.symbol ?? "").trim() ? normalizeContractSymbol(request.symbol) : undefined;
    const order = await exchange.cancelOrder(request.order_id ?? "", symbol);
    return { order: mapOrder(order) };
  }

  async setLeverage(request: SetLeverageRequest): Promise<{ ok: boolean }> {
    const { exchange } = await this.open(request.credentials);
    const leverage = toNumber(request.leverage);
    if (leverage <= 0) {
      return { ok: true };
    }
    await exchange.setLeverage(leverage, normalizeContractSymbol(request.symbol));
    return { ok: true };
  }

  async setMarginMode(request: SetMarginModeRequest): Promise<{ ok: boolean }> {
    const { exchange } = await this.open(request.credentials);
    const marginMode = request.is_cross_margin ? "cross" : "isolated";
    await exchange.setMarginMode(marginMode, normalizeContractSymbol(request.symbol));
    return { ok: true };
  }

  async cancelAllOrders(request: CancelAllOrdersRequest): Promise<{ ok: boolean }> {
    const { exchange } = await this.open(request.credentials);
    await exchange.cancelAllOrders(normalizeContractSymbol(request.symbol));
    return { ok: true };
  }

  async fetchOrder(request: FetchOrderRequest): Promise<{ order: ReturnType<typeof mapOrder> }> {
    const { exchange } = await this.open(request.credentials);
    const order = await exchange.fetchOrder(request.order_id ?? "", normalizeContractSymbol(request.symbol));
    return { order: mapOrder(order) };
  }

  async fetchOpenOrders(request: FetchOpenOrdersRequest): Promise<{ orders: ReturnType<typeof mapOrder>[] }> {
    const { exchange } = await this.open(request.credentials);
    const rows = await exchange.fetchOpenOrders(normalizeContractSymbol(request.symbol));
    return { orders: rows.map(mapOrder) };
  }

  async fetchMyTrades(request: FetchMyTradesRequest): Promise<{ trades: ReturnType<typeof mapTrade>[] }> {
    const { exchange } = await this.open(request.credentials);
    const params = parseParamsJson(request.params_json);
    const rows = await exchange.fetchMyTrades(undefined, this.optionalSince(request.since_ms), this.optionalLimit(request.limit), params);
    return { trades: rows.map(mapTrade) };
  }

  async fetchClosedOrders(request: FetchClosedOrdersRequest): Promise<{ orders: ReturnType<typeof mapOrder>[] }> {
    const { exchange } = await this.open(request.credentials);
    const params = parseParamsJson(request.params_json);
    const rows = await exchange.fetchClosedOrders(undefined, this.optionalSince(request.since_ms), this.optionalLimit(request.limit), params);
    return { orders: rows.map(mapOrder) };
  }

  async fetchPositionsHistory(request: FetchPositionsHistoryRequest): Promise<{ positions: ReturnType<typeof mapPosition>[] }> {
    const context = await this.open(request.credentials);
    if (context.exchangeType !== EXCHANGE_TYPE_BYBIT) {
      throw new Error(`交易所 ${context.exchangeType} 不支持仓位历史`);
    }
    const target = context.exchange as RuntimeExchange & {
      // 这里只给 Bybit 的扩展能力补类型，避免污染通用交易所接口。
      fetchPositionsHistory?: (symbol?: string, since?: number, limit?: number, params?: Record<string, unknown>) => Promise<Position[]>;
    };
    if (!target.fetchPositionsHistory) {
      throw new Error(`交易所 ${context.exchangeType} 不支持仓位历史`);
    }
    const params = parseParamsJson(request.params_json);
    const rows = await target.fetchPositionsHistory(undefined, this.optionalSince(request.since_ms), this.optionalLimit(request.limit), params);
    return { positions: rows.map(mapPosition) };
  }

  private async open(credentials: ExchangeCredentials | undefined): Promise<ExchangeContext> {
    return createPrivateExchange(credentials, this.proxy);
  }

  private positionParams(exchangeType: string): Record<string, unknown> | undefined {
    if (exchangeType !== EXCHANGE_TYPE_BYBIT) {
      return undefined;
    }
    return {
      category: "linear",
      settleCoin: "USDT",
    };
  }

  private optionalSince(raw?: number | string): number | undefined {
    const current = toNumber(raw);
    return current > 0 ? current : undefined;
  }

  private optionalLimit(raw?: number | string): number | undefined {
    const current = toNumber(raw);
    return current > 0 ? current : undefined;
  }

  private readBalanceValue(values: Record<string, number> | undefined, asset: string): number {
    if (!values) {
      return 0;
    }
    return toNumber(values[asset]);
  }
}
