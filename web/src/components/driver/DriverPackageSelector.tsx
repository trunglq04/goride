import React from "react";
import { PackagesMeta } from "@/components/rider/PackagesMeta";
import { CarPackageSlug } from "@/types";
import { cn } from "@/utils/cn";
import { Users } from "lucide-react";

interface DriverPackageSelectorProps {
  onSelect: (packageSlug: CarPackageSlug) => void;
}

export function DriverPackageSelector({ onSelect }: DriverPackageSelectorProps) {
  return (
    <div className="flex items-center justify-center min-h-[calc(100vh-64px)] p-4 bg-[#f4f4f6]">
      <div className="bg-white border border-black/10 rounded-3xl p-6 sm:p-8 max-w-md w-full shadow-sm">
        <h2 className="font-brand text-2xl font-extrabold text-[#18181b] tracking-tight mb-1">
          Select Vehicle Class
        </h2>
        <p className="text-xs text-zinc-500 mb-6">
          Specify the vehicle class you will be driving today on GoRide.
        </p>

        <div className="space-y-3">
          {Object.entries(PackagesMeta).map(([slug, meta]) => (
            <div
              key={slug}
              className={cn(
                "flex items-center justify-between p-4 rounded-2xl border border-black/8 bg-[#fafafb] transition-all cursor-pointer tactile-press",
                "hover:border-[#18181b] hover:bg-white hover:shadow-sm"
              )}
              onClick={() => onSelect(slug as CarPackageSlug)}
            >
              <div className="flex items-center gap-3.5">
                <div className="p-1 flex items-center justify-center">
                  {meta?.icon}
                </div>
                <div>
                  <h3 className="font-bold text-sm text-[#18181b]">{meta?.name}</h3>
                  <p className="text-xs text-zinc-500">{meta?.description}</p>
                </div>
              </div>

              <div className="flex items-center gap-1 text-xs text-zinc-400">
                <Users className="w-3.5 h-3.5" />
                <span>{meta.seats}</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
