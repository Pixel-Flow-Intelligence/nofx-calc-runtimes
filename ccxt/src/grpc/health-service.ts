import grpc from "@grpc/grpc-js";

type UnaryCall<Request, Response> = grpc.handleUnaryCall<Request, Response>;

type HealthCheckRequest = {
  service?: string;
};

type HealthCheckResponse = {
  status: number;
};

type WatchResponse = HealthCheckResponse;

export function buildHealthHandlers(): {
  Check: UnaryCall<HealthCheckRequest, HealthCheckResponse>;
  Watch: UnaryCall<HealthCheckRequest, WatchResponse>;
} {
  return {
    // 保持标准 grpc.health.v1.Health，保证 gateway 的健康探测无需修改。
    Check(call, callback): void {
      callback(null, { status: 1 });
    },
    Watch(call, callback): void {
      callback(null, { status: 1 });
    },
  };
}
