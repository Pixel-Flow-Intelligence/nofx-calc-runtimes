export type ExchangeCredentials = {
  exchange_id?: string;
  exchange_type?: string;
  api_key?: string;
  secret_key?: string;
  passphrase?: string;
  testnet?: boolean;
  is_demo?: boolean;
};

export type GetCurrentPriceRequest = {
  symbol?: string;
  exchange?: string;
};

export type FetchKlinesRequest = {
  exchange?: string;
  symbol?: string;
  timeframe?: string;
  limit?: number | string;
};

export type FetchOpenInterestRequest = {
  exchange?: string;
  symbol?: string;
};

export type FetchFundingRateRequest = {
  exchange?: string;
  symbol?: string;
};

export type FetchBalanceRequest = {
  credentials?: ExchangeCredentials;
};

export type FetchPositionsRequest = {
  credentials?: ExchangeCredentials;
  symbols?: string[];
};

export type FetchPositionRequest = {
  credentials?: ExchangeCredentials;
  symbol?: string;
};

export type FetchTickerRequest = {
  credentials?: ExchangeCredentials;
  symbol?: string;
};

export type FetchSymbolTradingRulesRequest = {
  exchange?: string;
  symbols?: string[];
};

export type PlaceMarketOrderRequest = {
  credentials?: ExchangeCredentials;
  symbol?: string;
  side?: string;
  amount?: number | string;
  params_json?: string;
};

export type PlaceLimitOrderRequest = {
  credentials?: ExchangeCredentials;
  symbol?: string;
  side?: string;
  amount?: number | string;
  price?: number | string;
  params_json?: string;
};

export type CancelOrderRequest = {
  credentials?: ExchangeCredentials;
  order_id?: string;
  symbol?: string;
};

export type SetLeverageRequest = {
  credentials?: ExchangeCredentials;
  symbol?: string;
  leverage?: number | string;
};

export type SetMarginModeRequest = {
  credentials?: ExchangeCredentials;
  symbol?: string;
  is_cross_margin?: boolean;
};

export type CancelAllOrdersRequest = {
  credentials?: ExchangeCredentials;
  symbol?: string;
};

export type FetchOrderRequest = {
  credentials?: ExchangeCredentials;
  order_id?: string;
  symbol?: string;
};

export type FetchOpenOrdersRequest = {
  credentials?: ExchangeCredentials;
  symbol?: string;
};

export type FetchMyTradesRequest = {
  credentials?: ExchangeCredentials;
  since_ms?: number | string;
  limit?: number | string;
  params_json?: string;
};

export type FetchClosedOrdersRequest = FetchMyTradesRequest;
export type FetchPositionsHistoryRequest = FetchMyTradesRequest;

export type KlineData = {
  open_time: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  close_time: number;
};

export type PositionData = {
  id: string;
  symbol: string;
  side: string;
  contracts: number;
  entry_price: number;
  mark_price: number;
  unrealized_pnl: number;
  leverage: number;
  realized_pnl: number;
  timestamp: number;
  info_json: string;
};

export type OrderData = {
  id: string;
  status: string;
  symbol: string;
  side: string;
  type: string;
  amount: number;
  filled: number;
  average: number;
  price: number;
  trigger_price: number;
  stop_loss_price: number;
  take_profit_price: number;
  timestamp: number;
  info_json: string;
  fee_cost: number;
  client_order_id: string;
};

export type TradeData = {
  id: string;
  order_id: string;
  symbol: string;
  side: string;
  type: string;
  amount: number;
  price: number;
  cost: number;
  fee_cost: number;
  fee_currency: string;
  maker: boolean;
  timestamp: number;
  info_json: string;
  realized_pnl: number;
  client_order_id: string;
  position_side: string;
  reduce_only: boolean;
};

export type TickerData = {
  last: number;
  close: number;
  bid: number;
  ask: number;
};
