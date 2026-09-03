"use client";

import React from "react";
import Link from "next/link";
import { Zap, ShieldCheck, Navigation } from "lucide-react";
import { BrandLogo } from "../BrandLogo";

interface AuthLayoutProps {
  children: React.ReactNode;
  title: string;
  subtitle: string;
}

export function AuthLayout({ children, title, subtitle }: AuthLayoutProps) {
  return (
    <div className="relative min-h-screen w-full bg-[#f4f4f6] text-[#18181b] flex items-center justify-center p-4 sm:p-6 lg:p-8">
      <div className="relative z-10 w-full max-w-5xl grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-12 items-center">
        {/* Left Column: Brand & Value Proposition (Desktop) */}
        <div className="hidden lg:flex lg:col-span-6 flex-col justify-between space-y-8 pr-4">
          <div>
            {/* Velocity Loop Brand Logo */}
            <BrandLogo size="lg" href="/" />

            {/* Main Headline */}
            <div className="mt-8 space-y-3">
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-zinc-100 border border-black/5 text-zinc-700 text-xs font-semibold">
                <span className="w-2 h-2 rounded-full bg-emerald-600 animate-pulse" />
                <span>Microservice Dispatch Network</span>
              </div>
              <h1 className="font-brand text-4xl xl:text-5xl font-extrabold tracking-tight text-[#18181b] leading-tight">
                Elevated transit, <br />
                built for clarity.
              </h1>
              <p className="text-zinc-500 text-sm max-w-md leading-relaxed">
                Connect seamlessly with verified chauffeurs, view upfront route fares, and experience calm, predictable urban mobility.
              </p>
            </div>
          </div>

          {/* Feature Badges */}
          <div className="space-y-3">
            <div className="flex items-center gap-4 p-3.5 rounded-2xl bg-white border border-black/6 shadow-sm">
              <div className="h-10 w-10 rounded-xl bg-zinc-100 flex items-center justify-center text-[#18181b] shrink-0">
                <Zap className="h-5 w-5" />
              </div>
              <div>
                <h4 className="text-xs font-bold text-[#18181b]">Sub-Second Driver Dispatch</h4>
                <p className="text-[11px] text-zinc-500">Real-time telemetry and geohashing match you with the closest car.</p>
              </div>
            </div>

            <div className="flex items-center gap-4 p-3.5 rounded-2xl bg-white border border-black/6 shadow-sm">
              <div className="h-10 w-10 rounded-xl bg-zinc-100 flex items-center justify-center text-emerald-700 shrink-0">
                <ShieldCheck className="h-5 w-5" />
              </div>
              <div>
                <h4 className="text-xs font-bold text-[#18181b]">Stripe Escrow Pre-Authorization</h4>
                <p className="text-[11px] text-zinc-500">Zero surprise charges. Payment is captured only upon completed arrival.</p>
              </div>
            </div>

            <div className="flex items-center gap-4 p-3.5 rounded-2xl bg-white border border-black/6 shadow-sm">
              <div className="h-10 w-10 rounded-xl bg-zinc-100 flex items-center justify-center text-[#18181b] shrink-0">
                <Navigation className="h-5 w-5" />
              </div>
              <div>
                <h4 className="text-xs font-bold text-[#18181b]">Transparent Geo-Routing</h4>
                <p className="text-[11px] text-zinc-500">Precision GPS routing and real-time step-by-step driver tracking.</p>
              </div>
            </div>
          </div>

          {/* Footer note */}
          <div className="text-[11px] text-zinc-400 flex items-center justify-between pt-4 border-t border-black/5">
            <span>© 2026 GoRide Transit Network.</span>
            <span className="flex items-center gap-1.5 text-emerald-700 font-semibold">
              <span className="h-1.5 w-1.5 rounded-full bg-emerald-600 animate-ping" />
              All Systems Operational
            </span>
          </div>
        </div>

        {/* Right Column: Form Card */}
        <div className="lg:col-span-6 w-full max-w-md mx-auto">
          {/* Mobile Brand */}
          <div className="lg:hidden flex items-center justify-center mb-6">
            <BrandLogo size="md" href="/" />
          </div>

          <div className="bg-white border border-black/10 rounded-3xl p-6 sm:p-8 shadow-md relative overflow-hidden">
            <div className="mb-6 text-center sm:text-left">
              <h2 className="font-brand text-2xl sm:text-3xl font-extrabold tracking-tight text-[#18181b]">
                {title}
              </h2>
              <p className="text-xs text-zinc-500 mt-1">{subtitle}</p>
            </div>

            {children}
          </div>
        </div>
      </div>
    </div>
  );
}
