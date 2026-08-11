import type { OHLCV, Order, Position, Ticker, Trade } from "ccxt";

import {
  stringifyInfo,
  toBoolean,
  toNumber,
  toStringValue,
} from "./helpers.js";
import type { KlineData, OrderData, PositionData, TickerData, TradeData } from "../types/proto.js";

function readInfoString(info: unknown, keys: string[]): string {
  if (!info || typeof info !== "object") {
    return "";
  }
  const payload = info as Record<string, unknown>;
  for (const key of keys) {
    const value = payload[key];
    if (typeof value === "string") {
      return value;
    }
  }
  return "";
}

function readInfoNumber(info: unknown, keys: string[]): number {
  if (!info || typeof info !== "object") {
    return 0;
  }
  const payload = info as Record<string, unknown>;
  for (const key of keys) {
    const value = payload[key];
    const casted = toNumber(value);
    if (casted !== 0) {
      return casted;
    }
  }
  return 0;
}

function readInfoBool(info: unknown, keys: string[]): boolean {
  if (!info || typeof info !== "object") {
    return false;
  }
  const payload = info as Record<string, unknown>;
  for (const key of keys) {
    if (toBoolean(payload[key])) {
      return true;
    }
  }
  return false;
}

export function mapKline(item: OHLCV, timeframeMs: number): KlineData {
  return {
    open_time: toNumber(item[0]),
    open: toNumber(item[1]),
    high: toNumber(item[2]),
    low: toNumber(item[3]),
    close: toNumber(item[4]),
    volume: toNumber(item[5]),
    close_time: toNumber(item[0]) + timeframeMs - 1,
  };
}

export function mapTicker(item: Ticker): TickerData {
  return {
    last: toNumber(item.last),
    close: toNumber(item.close),
    bid: toNumber(item.bid),
    ask: toNumber(item.ask),
  };
}

export function mapPosition(item: Position): PositionData {
  return {
    id: toStringValue(item.id),
    symbol: toStringValue(item.symbol),
    side: toStringValue(item.side),
    contracts: toNumber(item.contracts),
    entry_price: toNumber(item.entryPrice),
    mark_price: toNumber(item.markPrice),
    unrealized_pnl: toNumber(item.unrealizedPnl),
    leverage: toNumber(item.leverage),
    realized_pnl: toNumber(item.realizedPnl),
    timestamp: toNumber(item.timestamp),
    info_json: stringifyInfo(item.info),
  };
}

export function mapOrder(item: Order): OrderData {
  return {
    id: toStringValue(item.id),
    status: toStringValue(item.status),
    symbol: toStringValue(item.symbol),
    side: toStringValue(item.side),
    type: toStringValue(item.type),
    amount: toNumber(item.amount),
    filled: toNumber(item.filled),
    average: toNumber(item.average),
    price: toNumber(item.price),
    trigger_price: toNumber(item.triggerPrice),
    stop_loss_price: toNumber(item.stopLossPrice),
    take_profit_price: toNumber(item.takeProfitPrice),
    timestamp: toNumber(item.timestamp),
    info_json: stringifyInfo(item.info),
    fee_cost: toNumber(item.fee?.cost),
    client_order_id: toStringValue(item.clientOrderId),
  };
}

export function mapTrade(item: Trade): TradeData {
  return {
    id: toStringValue(item.id),
    order_id: toStringValue(item.order),
    symbol: toStringValue(item.symbol),
    side: toStringValue(item.side),
    type: toStringValue(item.type),
    amount: toNumber(item.amount),
    price: toNumber(item.price),
    cost: toNumber(item.cost),
    fee_cost: toNumber(item.fee?.cost),
    fee_currency: toStringValue(item.fee?.currency),
    maker: toStringValue(item.takerOrMaker) === "maker",
    timestamp: toNumber(item.timestamp),
    info_json: stringifyInfo(item.info),
    realized_pnl: readInfoNumber(item.info, ["realizedPnl", "closedPnl", "fillPnl", "execPnl", "profit", "pnl"]),
    client_order_id: readInfoString(item.info, ["clientOrderId", "clientOid", "clOrdId", "orderLinkId", "c"]),
    position_side: readInfoString(item.info, ["positionSide", "posSide", "tradeSide"]),
    reduce_only: readInfoBool(item.info, ["reduceOnly", "reduce_only", "closePosition", "close_position"]),
  };
}
