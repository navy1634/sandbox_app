"use client";

import { createContext, type ReactNode, useContext } from "react";

import type { RuntimeConfig } from "./runtimeConfig";

const RuntimeConfigContext = createContext<RuntimeConfig | undefined>(
  undefined,
);

export function RuntimeConfigProvider({
  children,
  value,
}: Readonly<{
  children: ReactNode;
  value: RuntimeConfig;
}>) {
  return <RuntimeConfigContext value={value}>{children}</RuntimeConfigContext>;
}

export function useRuntimeConfig() {
  const runtimeConfig = useContext(RuntimeConfigContext);

  if (!runtimeConfig) {
    throw new Error("RuntimeConfigProvider is required");
  }

  return runtimeConfig;
}
