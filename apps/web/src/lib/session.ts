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
