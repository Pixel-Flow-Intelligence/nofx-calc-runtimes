import dotenv from "dotenv";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadConfig } from "./config.js";
import { logger } from "./logger.js";
import { startGrpcServer } from "./server.js";

const CURRENT_DIR = path.dirname(fileURLToPath(import.meta.url));

dotenv.config();
dotenv.config({ path: path.resolve(CURRENT_DIR, "../../.env") });

async function main(): Promise<void> {
  const config = loadConfig();
  await startGrpcServer(config);
}

main().catch((error) => {
  logger.error("ccxt-runtime 异常退出", { error: String(error) });
  process.exit(1);
});
