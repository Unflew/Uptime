import { NextRequest } from "next/server";
import { ACCESS_COOKIE } from "../../../../lib/auth";

export async function GET(request: NextRequest, { params }: { params: Promise<{ name: string }> }) {
  const { name } = await params;
  if (!/^[a-f0-9]{32}$/.test(name)) return new Response(null, { status: 404 });
  const token = request.cookies.get(ACCESS_COOKIE)?.value;
  if (!token) return new Response(null, { status: 401 });
  try {
    const upstream = await fetch(`${process.env.BACKEND_URL ?? "http://127.0.0.1:8080"}/files/${name}`, { headers: { Authorization: `Bearer ${token}` }, cache: "no-store", signal: AbortSignal.timeout(10000) });
    if (!upstream.ok) return new Response(null, { status: upstream.status });
    return new Response(upstream.body, { headers: { "Content-Type": "application/octet-stream", "Content-Disposition": "attachment", "X-Content-Type-Options": "nosniff", "Cache-Control": "no-store" } });
  } catch { return new Response(null, { status: 503 }); }
}
