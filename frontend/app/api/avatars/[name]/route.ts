import { NextRequest } from "next/server";

export async function GET(_request: NextRequest, { params }: { params: Promise<{ name: string }> }) {
  const { name } = await params;
  if (!/^[a-f0-9]{32}\.(jpeg|png|gif)$/.test(name)) return new Response(null, { status: 404 });
  try {
    const upstream = await fetch(`${process.env.BACKEND_URL ?? "http://127.0.0.1:8080"}/avatars/${name}`, { signal: AbortSignal.timeout(10000) });
    if (!upstream.ok) return new Response(null, { status: upstream.status });
    return new Response(upstream.body, { headers: { "Content-Type": upstream.headers.get("content-type") ?? "application/octet-stream", "X-Content-Type-Options": "nosniff", "Cache-Control": "public, max-age=31536000, immutable" } });
  } catch { return new Response(null, { status: 503 }); }
}
