import { cookies } from "next/headers";

export const SESSION_COOKIE = "talerole_session";

export async function getSessionToken(): Promise<string | undefined> {
  return (await cookies()).get(SESSION_COOKIE)?.value;
}

export function sessionCookieOptions() {
  return {
    httpOnly: true,
    sameSite: "lax" as const,
    secure: process.env.NODE_ENV === "production",
    path: "/",
    maxAge: 60 * 60 * 8,
  };
}

export function apiBase(): string {
  return process.env.API_URL ?? "http://127.0.0.1:8080";
}

export function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  return fetch(`${apiBase()}${path}`, {
    ...init,
    signal: init.signal ?? AbortSignal.timeout(20_000),
  });
}

export async function gqlUpstream<T>(
  token: string | undefined,
  query: string,
  variables?: Record<string, unknown>,
): Promise<{ data?: T; errors?: { message: string }[] }> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  const res = await fetch(`${apiBase()}/graphql`, {
    method: "POST",
    headers,
    body: JSON.stringify({ query, variables }),
    cache: "no-store",
  });
  return (await res.json().catch(() => ({}))) as { data?: T; errors?: { message: string }[] };
}
