import grpc from "@grpc/grpc-js";

import type {
  CancelAllOrdersRequest,
  CancelOrderRequest,
  FetchBalanceRequest,
  FetchClosedOrdersRequest,
  FetchFundingRateRequest,
  FetchKlinesRequest,
  FetchMyTradesRequest,
  FetchOpenInterestRequest,
  FetchOpenOrdersRequest,
  FetchOrderRequest,
  FetchPositionRequest,
  FetchPositionsHistoryRequest,
  FetchPositionsRequest,
  FetchSymbolTradingRulesRequest,
  FetchTickerRequest,
  GetCurrentPriceRequest,
  PlaceLimitOrderRequest,
  PlaceMarketOrderRequest,
  SetLeverageRequest,
  SetMarginModeRequest,
} from "../types/proto.js";
import { logger } from "../logger.js";
import { PublicMarketService } from "../services/public-service.js";
import { PrivateExchangeService } from "../services/private-service.js";

type UnaryHandler<Request, Response> = grpc.handleUnaryCall<Request, Response>;

function wrapUnary<Request, Response>(
  methodName: string,
  handler: (request: Request) => Promise<Response>,
): UnaryHandler<Request, Response> {
  return async (call, callback): Promise<void> => {
    try {
      logger.info("收到 gRPC 请求", { method: methodName });
      const response = await handler(call.request);
      callback(null, response);
    } catch (error) {
      logger.error("gRPC 请求失败", { method: methodName, error: String(error) });
      callback({
        code: grpc.status.INTERNAL,
        message: error instanceof Error ? error.message : String(error),
      });
    }
  };
}

export function buildCcxtRuntimeHandlers(
  publicService: PublicMarketService,
  privateService: PrivateExchangeService,
): grpc.UntypedServiceImplementation {
  return {
    GetCurrentPrice: wrapUnary<GetCurrentPriceRequest, { price: number }>("GetCurrentPrice", (request) =>
      publicService.getCurrentPrice(request),
    ),
    FetchKlines: wrapUnary<FetchKlinesRequest, { klines: unknown[] }>("FetchKlines", (request) =>
      publicService.fetchKlines(request),
    ),
    FetchOpenInterest: wrapUnary<FetchOpenInterestRequest, { latest: number; average: number }>("FetchOpenInterest", (request) =>
      publicService.fetchOpenInterest(request),
    ),
    FetchFundingRate: wrapUnary<FetchFundingRateRequest, { rate: number }>("FetchFundingRate", (request) =>
      publicService.fetchFundingRate(request),
    ),
    FetchSymbolTradingRules: wrapUnary<FetchSymbolTradingRulesRequest, { rules: unknown[] }>("FetchSymbolTradingRules", (request) =>
      publicService.fetchSymbolTradingRules(request),
    ),
    FetchBalance: wrapUnary<FetchBalanceRequest, { available_balance: number; margin_used: number; wallet_balance: number }>(
      "FetchBalance",
      (request) => privateService.fetchBalance(request),
    ),
    FetchPositions: wrapUnary<FetchPositionsRequest, { positions: unknown[] }>("FetchPositions", (request) =>
      privateService.fetchPositions(request),
    ),
    FetchPosition: wrapUnary<FetchPositionRequest, { position: unknown }>("FetchPosition", (request) =>
      privateService.fetchPosition(request),
    ),
    FetchTicker: wrapUnary<FetchTickerRequest, { ticker: unknown }>("FetchTicker", (request) =>
      privateService.fetchTicker(request),
    ),
    PlaceMarketOrder: wrapUnary<PlaceMarketOrderRequest, { order: unknown }>("PlaceMarketOrder", (request) =>
      privateService.placeMarketOrder(request),
    ),
    PlaceLimitOrder: wrapUnary<PlaceLimitOrderRequest, { order: unknown }>("PlaceLimitOrder", (request) =>
      privateService.placeLimitOrder(request),
    ),
    CancelOrder: wrapUnary<CancelOrderRequest, { order: unknown }>("CancelOrder", (request) =>
      privateService.cancelOrder(request),
    ),
    SetLeverage: wrapUnary<SetLeverageRequest, { ok: boolean }>("SetLeverage", (request) =>
      privateService.setLeverage(request),
    ),
    SetMarginMode: wrapUnary<SetMarginModeRequest, { ok: boolean }>("SetMarginMode", (request) =>
      privateService.setMarginMode(request),
    ),
    CancelAllOrders: wrapUnary<CancelAllOrdersRequest, { ok: boolean }>("CancelAllOrders", (request) =>
      privateService.cancelAllOrders(request),
    ),
    FetchOrder: wrapUnary<FetchOrderRequest, { order: unknown }>("FetchOrder", (request) =>
      privateService.fetchOrder(request),
    ),
    FetchOpenOrders: wrapUnary<FetchOpenOrdersRequest, { orders: unknown[] }>("FetchOpenOrders", (request) =>
      privateService.fetchOpenOrders(request),
    ),
    FetchMyTrades: wrapUnary<FetchMyTradesRequest, { trades: unknown[] }>("FetchMyTrades", (request) =>
      privateService.fetchMyTrades(request),
    ),
    FetchClosedOrders: wrapUnary<FetchClosedOrdersRequest, { orders: unknown[] }>("FetchClosedOrders", (request) =>
      privateService.fetchClosedOrders(request),
    ),
    FetchPositionsHistory: wrapUnary<FetchPositionsHistoryRequest, { positions: unknown[] }>("FetchPositionsHistory", (request) =>
      privateService.fetchPositionsHistory(request),
    ),
  };
}
