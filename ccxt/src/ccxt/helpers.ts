export function normalizeExchangeType(exchange?: string): string {
  return (exchange ?? "").trim().toLowerCase();
}

export function normalizeSpotSymbol(symbol?: string): string {
  const current = (symbol ?? "").trim().toUpperCase();
  if (!current) {
    return current;
  }
  if (current.includes("/")) {
    return current;
  }
  if (current.endsWith("USDT")) {
    return `${current.slice(0, -4)}/USDT`;
  }
  return `${current}/USDT`;
}

export function normalizeContractSymbol(symbol?: string): string {
  const current = normalizeSpotSymbol(symbol);
  if (!current || current.includes(":")) {
    return current;
  }
  if (current.endsWith("/USDT")) {
    return `${current}:USDT`;
  }
  return current;
}

export function timeframeToMillis(timeframe?: string): number {
  switch ((timeframe ?? "").trim()) {
    case "1m":
      return 60_000;
    case "5m":
      return 300_000;
    case "15m":
      return 900_000;
    case "30m":
      return 1_800_000;
    case "1h":
      return 3_600_000;
    case "4h":
      return 14_400_000;
    case "1d":
      return 86_400_000;
    default:
      return 60_000;
  }
}

export function parseParamsJson(raw?: string): Record<string, unknown> | undefined {
  const current = (raw ?? "").trim();
  if (!current) {
    return undefined;
  }
  return JSON.parse(current) as Record<string, unknown>;
}

export function stringifyInfo(value: unknown): string {
  if (value == null) {
    return "";
  }
  try {
    return JSON.stringify(value);
  } catch {
    return "";
  }
}

export function toNumber(value: unknown): number {
  if (typeof value === "number") {
    return Number.isFinite(value) ? value : 0;
  }
  if (typeof value === "bigint") {
    return Number(value);
  }
  if (typeof value === "string") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

export function toBoolean(value: unknown): boolean {
  return value === true;
}

export function toStringValue(value: unknown): string {
  return typeof value === "string" ? value : "";
}
