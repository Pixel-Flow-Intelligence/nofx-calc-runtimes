import path from "node:path";

import grpc from "@grpc/grpc-js";
import protoLoader from "@grpc/proto-loader";

const LOADER_OPTIONS: protoLoader.Options = {
  longs: Number,
  enums: String,
  defaults: true,
  oneofs: true,
  keepCase: true,
};

export type LoadedProto = {
  nofx: {
    ccxtruntime: {
      v1: {
        CCXTRuntimeService: {
          service: grpc.ServiceDefinition;
        };
      };
    };
  };
  grpc: {
    health: {
      v1: {
        Health: {
          service: grpc.ServiceDefinition;
        };
      };
    };
  };
};

export function loadProto(protoPath: string, healthProtoPath: string): LoadedProto {
  const packageDefinition = protoLoader.loadSync([protoPath, healthProtoPath], {
    ...LOADER_OPTIONS,
    includeDirs: [
      path.resolve(path.dirname(protoPath), "../../../../"),
      path.resolve(path.dirname(healthProtoPath), "../../../"),
    ],
  });
  return grpc.loadPackageDefinition(packageDefinition) as unknown as LoadedProto;
}
