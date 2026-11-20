import type { Metadata } from "next";
import type { ReactNode } from "react";
import "./styles.css";

export const metadata: Metadata = {
  title: "Вход — Uptime Up",
  description: "Вход и регистрация в Uptime Up",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return <html lang="ru"><body>{children}</body></html>;
}
