type LogPayload = Record<string, unknown>;

function formatPayload(payload?: LogPayload): string {
  if (!payload || Object.keys(payload).length === 0) {
    return "";
  }
  return ` ${JSON.stringify(payload)}`;
}

export const logger = {
  info(message: string, payload?: LogPayload): void {
    console.log(`[ccxt] ${message}${formatPayload(payload)}`);
  },
  error(message: string, payload?: LogPayload): void {
    console.error(`[ccxt] ${message}${formatPayload(payload)}`);
  },
};
