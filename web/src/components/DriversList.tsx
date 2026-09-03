import React, { useState } from "react";
import { Button } from "./ui/button";
import { Clock, Users, ArrowRight, Shield } from "lucide-react";
import { RouteFare, TripPreview } from "../types";
import {
  convertMetersToKilometers,
  convertSecondsToMinutes,
} from "../utils/math";
import { cn } from "../lib/utils";
import { PackagesMeta } from "./PackagesMeta";

interface DriverListProps {
  trip: TripPreview | null;
  onPackageSelect: (fare: RouteFare) => void;
  onCancel: () => void;
}

export function DriverList({
  trip,
  onPackageSelect,
  onCancel,
}: DriverListProps) {
  const rideFares = trip?.rideFares ?? [];
  const [selectedFareId, setSelectedFareId] = useState<string>(
    rideFares[0]?.id ?? ""
  );

  const selectedFare =
    rideFares.find((f) => f.id === selectedFareId) || rideFares[0];

  const selectedMeta = selectedFare
    ? PackagesMeta[selectedFare.packageSlug]
    : null;

  const formattedSelectedPrice = selectedFare?.totalPriceInCents
    ? `$${(selectedFare.totalPriceInCents / 100).toFixed(2)}`
    : "$0.00";

  return (
    <div className="w-full flex flex-col gap-4">
      {/* Route Telemetry Strip */}
      <div className="flex items-center justify-between px-4 py-2.5 bg-[#fafafb] border border-black/5 rounded-2xl text-xs">
        <div className="flex items-center gap-2">
          <Clock className="w-3.5 h-3.5 text-zinc-500" />
          <span className="font-semibold text-zinc-800">
            {convertSecondsToMinutes(trip?.duration ?? 0)} duration
          </span>
        </div>
        <span className="text-zinc-400">•</span>
        <span className="font-medium text-zinc-600">
          {convertMetersToKilometers(trip?.distance ?? 0)} distance
        </span>
        <span className="text-zinc-400">•</span>
        <div className="flex items-center gap-1 text-emerald-700 font-semibold">
          <Shield className="w-3.5 h-3.5" />
          <span>Optimal Route</span>
        </div>
      </div>

      {/* Ride Tiers Stack */}
      <div className="space-y-2.5 max-h-[300px] overflow-y-auto pr-0.5">
        {rideFares.map((fare) => {
          const meta = PackagesMeta[fare.packageSlug];
          const isSelected = fare.id === selectedFareId;
          const price = fare.totalPriceInCents
            ? `$${(fare.totalPriceInCents / 100).toFixed(2)}`
            : "—";

          return (
            <div
              key={fare.id}
              onClick={() => setSelectedFareId(fare.id)}
              className={cn(
                "flex items-center justify-between p-3.5 rounded-2xl border transition-all cursor-pointer tactile-press",
                isSelected
                  ? "bg-white border-[#18181b] shadow-md ring-1 ring-[#18181b]"
                  : "bg-[#fafafb] border-black/5 hover:border-black/20 hover:bg-white"
              )}
            >
              <div className="flex items-center gap-3.5">
                {/* SVG Silhouette */}
                <div className="p-1 flex items-center justify-center">
                  {meta.icon}
                </div>

                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="font-bold text-sm text-[#18181b]">
                      {meta.name}
                    </h3>
                    {meta.badge && (
                      <span className="text-[10px] font-semibold px-2 py-0.5 rounded-full bg-zinc-100 text-zinc-700 border border-black/5">
                        {meta.badge}
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-2 text-xs text-zinc-500 mt-0.5">
                    <span className="flex items-center gap-1">
                      <Users className="w-3 h-3 text-zinc-400" />
                      {meta.seats} seats
                    </span>
                    <span>•</span>
                    <span>{meta.description}</span>
                  </div>
                </div>
              </div>

              {/* Price Display */}
              <div className="text-right">
                <p className="font-serif-brand font-bold text-base text-[#18181b]">
                  {price}
                </p>
              </div>
            </div>
          );
        })}
      </div>

      {/* Confirm Button */}
      <div className="pt-1 flex flex-col gap-2">
        <Button
          onClick={() => selectedFare && onPackageSelect(selectedFare)}
          className="w-full py-6 rounded-2xl bg-[#18181b] hover:bg-zinc-800 text-white font-semibold text-sm shadow-md flex items-center justify-center gap-2 tactile-press cursor-pointer"
        >
          <span>
            Confirm {selectedMeta?.name ?? "Ride"} • {formattedSelectedPrice}
          </span>
          <ArrowRight className="w-4 h-4" />
        </Button>

        <button
          type="button"
          onClick={onCancel}
          className="text-xs font-semibold text-zinc-500 hover:text-zinc-800 py-2 transition-colors cursor-pointer text-center"
        >
          Cancel & Choose Another Destination
        </button>
      </div>
    </div>
  );
}
