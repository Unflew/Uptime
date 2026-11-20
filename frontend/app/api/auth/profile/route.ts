import { NextRequest, NextResponse } from "next/server";
import { ACCESS_COOKIE } from "../../../../lib/auth";

async function proxy(request: NextRequest, method: "GET" | "PATCH") {
  if (method === "PATCH") {
    const origin = request.headers.get("origin");
    if (origin) {
      try {
        const url = new URL(origin);
        if (url.host !== request.headers.get("host") || url.protocol !== request.nextUrl.protocol) return NextResponse.json({ error: "Invalid origin" }, { status: 403 });
      } catch { return NextResponse.json({ error: "Invalid origin" }, { status: 403 }); }
    }
  }
  const token = request.cookies.get(ACCESS_COOKIE)?.value;
  if (!token) return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  try {
    const upstream = await fetch(`${process.env.BACKEND_URL ?? "http://127.0.0.1:8080"}/auth/me`, {
      method,
      headers: { Authorization: `Bearer ${token}`, ...(method === "PATCH" ? { "Content-Type": "application/json" } : {}) },
      ...(method === "PATCH" ? { body: await request.text() } : {}),
      cache: "no-store",
      signal: AbortSignal.timeout(10000),
    });
    if (!upstream.ok) return NextResponse.json({ error: "Unable to load profile" }, { status: upstream.status });
    return NextResponse.json(await upstream.json(), { headers: { "Cache-Control": "no-store" } });
  } catch {
    return NextResponse.json({ error: "Service unavailable" }, { status: 503 });
  }
}

export async function GET(request: NextRequest) { return proxy(request, "GET"); }
export async function PATCH(request: NextRequest) { return proxy(request, "PATCH"); }
