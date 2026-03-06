import type { ApiResponse } from "@/api/types";

const API_BASE = "/api/v1";

export class ApiRequestError extends Error {
  code: string;
  status: number;

  constructor(message: string, options: { code: string; status: number }) {
    super(message);
    this.code = options.code;
    this.status = options.status;
  }
}

type RequestOptions = RequestInit & {
  timeoutMs?: number;
};

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { timeoutMs = 10_000, ...requestOptions } = options;

  try {
    const response = await Promise.race([
      fetch(`${API_BASE}${path}`, {
        ...requestOptions,
        headers: {
          Accept: "application/json",
          ...(requestOptions.body instanceof FormData ? {} : { "Content-Type": "application/json" }),
          ...requestOptions.headers,
        },
      }),
      new Promise<Response>((_, reject) => {
        setTimeout(() => reject(new ApiRequestError("request timeout", { code: "TIMEOUT", status: 408 })), timeoutMs);
      }),
    ]);

    const payload = (await response.json()) as ApiResponse<T>;

    if (!response.ok || payload.error) {
      throw new ApiRequestError(payload.error?.message ?? "request failed", {
        code: payload.error?.code ?? "REQUEST_FAILED",
        status: response.status,
      });
    }

    if (payload.data === null) {
      throw new ApiRequestError("response data is null", {
        code: "EMPTY_DATA",
        status: response.status,
      });
    }

    return payload.data;
  } finally {
    // Promise.race timeout uses a detached timer; no abort cleanup needed.
  }
}
