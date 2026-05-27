const API_BASE = import.meta.env.VITE_API_BASE ?? "http://127.0.0.1:3001/api";
const API_KEY = import.meta.env.VITE_API_KEY ?? "dev-secret";

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number
  ) {
    super(message);
  }
}

export async function apiRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("x-api-key", API_KEY);

  if (!(init.body instanceof FormData)) {
    headers.set("content-type", "application/json");
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers
  });

  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }));
    throw new ApiError(body.error ?? "Request failed", response.status);
  }

  return response.json() as Promise<T>;
}

export function downloadUrl(path: string): string {
  return `${API_BASE}${path}`;
}

export function apiKey(): string {
  return API_KEY;
}

