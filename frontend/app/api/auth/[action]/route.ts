import { NextRequest, NextResponse } from "next/server";
import { ACCESS_COOKIE, REFRESH_COOKIE } from "../../../../lib/auth";

const backendUrl = process.env.BACKEND_URL ?? "http://127.0.0.1:8080";

export async function POST(request: NextRequest, context: { params: Promise<{ action: string }> }) {
  const { action } = await context.params;
  if (!["login", "register", "refresh", "logout"].includes(action)) {
    return NextResponse.json({ error: "Not found" }, { status: 404 });
  }
  const origin = request.headers.get("origin");
  if (origin && origin !== request.nextUrl.origin) {
    return NextResponse.json({ error: "Invalid origin" }, { status: 403 });
  }
  if (action === "logout") {
    const response = NextResponse.json({ ok: true });
    response.cookies.delete(ACCESS_COOKIE);
    response.cookies.delete(REFRESH_COOKIE);
    return response;
  }

  let body: unknown;
  if (action === "refresh") {
    const refreshToken = request.cookies.get(REFRESH_COOKIE)?.value;
    if (!refreshToken) return NextResponse.json({ error: "Session expired" }, { status: 401 });
    body = { refresh_token: refreshToken };
  } else {
    try {
      body = await request.json();
    } catch {
      return NextResponse.json({ error: "Invalid request" }, { status: 400 });
    }
  }

  try {
    const upstream = await fetch(`${backendUrl}/auth/${action}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
      cache: "no-store",
      signal: AbortSignal.timeout(10000),
    });
    if (!upstream.ok) {
      const response = NextResponse.json({ error: "Authentication failed" }, { status: upstream.status });
      if (action === "refresh" && upstream.status === 401) {
        response.cookies.delete(ACCESS_COOKIE);
        response.cookies.delete(REFRESH_COOKIE);
      }
      return response;
    }
    const tokens = await upstream.json();
    if (typeof tokens.access_token !== "string" || typeof tokens.refresh_token !== "string" ||
        !Number.isFinite(tokens.expires_in) || tokens.expires_in <= 0) {
      throw new Error("Invalid backend response");
    }
    const response = NextResponse.json({ ok: true }, { status: upstream.status });
    const cookieOptions = { httpOnly: true, secure: process.env.NODE_ENV === "production", sameSite: "lax" as const, path: "/" };
    response.cookies.set(ACCESS_COOKIE, tokens.access_token, { ...cookieOptions, maxAge: tokens.expires_in });
    // The backend is authoritative for refresh expiry.
    const configuredAge = Number(process.env.REFRESH_COOKIE_MAX_AGE ?? 30 * 24 * 60 * 60);
    const refreshMaxAge = Number.isSafeInteger(configuredAge) && configuredAge > 0 ? configuredAge : 30 * 24 * 60 * 60;
    response.cookies.set(REFRESH_COOKIE, tokens.refresh_token, { ...cookieOptions, maxAge: refreshMaxAge });
    return response;
  } catch {
    return NextResponse.json({ error: "Service unavailable" }, { status: 503 });
  }
}
