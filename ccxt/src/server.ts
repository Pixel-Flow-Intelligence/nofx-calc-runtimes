import grpc from "@grpc/grpc-js";

import type { RuntimeConfig } from "./config.js";
import { logger } from "./logger.js";
import { PublicExchangeFactory } from "./ccxt/exchange-factory.js";
import { buildHealthHandlers } from "./grpc/health-service.js";
import { loadProto } from "./grpc/proto-loader.js";
import { buildCcxtRuntimeHandlers } from "./grpc/service-handlers.js";
import { PrivateExchangeService } from "./services/private-service.js";
import { PublicMarketService } from "./services/public-service.js";

export async function startGrpcServer(config: RuntimeConfig): Promise<void> {
  const proto = loadProto(config.protoPath, config.healthProtoPath);
  const server = new grpc.Server();

  const publicFactory = new PublicExchangeFactory(config.ccxtProxy);
  const publicService = new PublicMarketService(publicFactory);
  const privateService = new PrivateExchangeService(config.ccxtProxy);

  server.addService(
    proto.nofx.ccxtruntime.v1.CCXTRuntimeService.service,
    buildCcxtRuntimeHandlers(publicService, privateService),
  );
  server.addService(proto.grpc.health.v1.Health.service, buildHealthHandlers());

  await new Promise<void>((resolve, reject) => {
    server.bindAsync(
      config.grpcListen,
      grpc.ServerCredentials.createInsecure(),
      (error, port) => {
        if (error) {
          reject(error);
          return;
        }
        server.start();
        logger.info("ccxt-runtime 已启动", {
          grpc_listen: config.grpcListen,
          port,
        });
        resolve();
      },
    );
  });
}
