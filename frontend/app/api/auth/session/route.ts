import { NextRequest, NextResponse } from "next/server";
import { ACCESS_COOKIE, REFRESH_COOKIE } from "../../../../lib/auth";

export async function GET(request: NextRequest) {
  return NextResponse.json({
    authenticated: Boolean(request.cookies.get(ACCESS_COOKIE)?.value),
    canRefresh: Boolean(request.cookies.get(REFRESH_COOKIE)?.value),
  }, { headers: { "Cache-Control": "no-store" } });
}
