import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import Footer from "@/components/Footer/components";
import Header from "@/components/Header/components";
import type { NavbarItem } from "@/components/Navbar/components";
import { createRuntimeConfig } from "./runtimeConfig";
import { RuntimeConfigProvider } from "./runtimeConfigContext";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "SSO Sandbox App",
  description: "sandbox_auth SSO を試すための簡易アプリです",
};

export const dynamic = "force-dynamic";

const navItems: NavbarItem[] = [
  { href: "/", label: "ホーム" },
  { href: "/dashboard", label: "ダッシュボード" },
];

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const runtimeConfig = createRuntimeConfig(process.env);

  return (
    <html lang="ja" className={`${geistSans.variable} ${geistMono.variable}`}>
      <body>
        <RuntimeConfigProvider value={runtimeConfig}>
          <Header navItems={navItems} /> {/* Headerコンポーネント*/}
          <main className="p-4">{children}</main>
          <Footer /> {/* Footerコンポーネント*/}
        </RuntimeConfigProvider>
      </body>
    </html>
  );
}
