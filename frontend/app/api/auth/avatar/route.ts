import { NextRequest, NextResponse } from "next/server";
import { ACCESS_COOKIE } from "../../../../lib/auth";

export async function POST(request: NextRequest) {
  const origin = request.headers.get("origin");
  if (origin) {
    try {
      const url = new URL(origin);
      if (url.host !== request.headers.get("host") || url.protocol !== request.nextUrl.protocol) return NextResponse.json({ error: "Invalid origin" }, { status: 403 });
    } catch { return NextResponse.json({ error: "Invalid origin" }, { status: 403 }); }
  }
  const token = request.cookies.get(ACCESS_COOKIE)?.value;
  if (!token) return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  try {
    const form = await request.formData();
    const file = form.get("avatar");
    if (!(file instanceof File) || file.size > 5 * 1024 * 1024) return NextResponse.json({ error: "Invalid image" }, { status: 400 });
    const body = new FormData();
    body.set("avatar", file);
    const upstream = await fetch(`${process.env.BACKEND_URL ?? "http://127.0.0.1:8080"}/auth/avatar`, {
      method: "POST", headers: { Authorization: `Bearer ${token}` }, body, signal: AbortSignal.timeout(10000),
    });
    return NextResponse.json(await upstream.json(), { status: upstream.status, headers: { "Cache-Control": "no-store" } });
  } catch { return NextResponse.json({ error: "Service unavailable" }, { status: 503 }); }
}
