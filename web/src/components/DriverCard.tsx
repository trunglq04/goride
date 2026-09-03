import React from "react";
import { Driver, CarPackageSlug } from "../types";
import Image from "next/image";
import { Phone, MessageSquare, ShieldCheck, Star } from "lucide-react";

interface DriverCardProps {
  driver?: Driver | null;
  packageSlug?: CarPackageSlug;
  etaMinutes?: number;
}

export const DriverCard = ({ driver, packageSlug, etaMinutes = 3 }: DriverCardProps) => {
  if (!driver) return null;

  const initials = driver.name
    ? driver.name
        .split(" ")
        .map((n) => n[0])
        .join("")
        .slice(0, 2)
        .toUpperCase()
    : "DR";

  return (
    <div className="w-full bg-white border border-black/10 rounded-2xl p-4 shadow-sm flex flex-col gap-3.5">
      {/* Top Pass Header: Avatar + Name + Metal Plate Badge */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="relative">
            {driver.profilePicture ? (
              <Image
                className="w-12 h-12 rounded-full object-cover border border-black/10 shadow-sm"
                src={driver.profilePicture}
                alt={`${driver.name}'s photo`}
                width={48}
                height={48}
              />
            ) : (
              <div className="w-12 h-12 rounded-full bg-[#18181b] text-white flex items-center justify-center font-bold text-sm shadow-sm">
                {initials}
              </div>
            )}
            <span className="absolute bottom-0 right-0 w-3.5 h-3.5 bg-emerald-600 border-2 border-white rounded-full shadow-sm" />
          </div>

          <div>
            <div className="flex items-center gap-1.5">
              <h3 className="text-base font-bold text-[#18181b] tracking-tight">{driver.name}</h3>
              <ShieldCheck className="w-4 h-4 text-emerald-600" />
            </div>
            <div className="flex items-center gap-1.5 text-xs text-zinc-500 mt-0.5">
              <div className="flex items-center gap-0.5 text-amber-600 font-semibold">
                <Star className="w-3 h-3 fill-amber-500 text-amber-500" />
                <span>4.98</span>
              </div>
              <span>•</span>
              <span>2,400+ rides</span>
              {packageSlug && (
                <>
                  <span>•</span>
                  <span className="capitalize text-zinc-700 font-medium">{packageSlug}</span>
                </>
              )}
            </div>
          </div>
        </div>

        {driver.carPlate && (
          <div className="plate-badge-metal text-xs tracking-wider">
            {driver.carPlate.toUpperCase()}
          </div>
        )}
      </div>

      {/* Arrival Milestone Banner */}
      <div className="flex items-center justify-between px-3 py-2 bg-[#fafafb] border border-black/5 rounded-xl text-xs">
        <span className="text-zinc-500 font-medium">Estimated arrival</span>
        <span className="font-bold text-emerald-700">~{etaMinutes} mins away</span>
      </div>

      {/* Quick Action Buttons */}
      <div className="grid grid-cols-2 gap-2 pt-1 border-t border-black/5">
        <button
          type="button"
          onClick={() => alert(`Calling driver at ${driver.name}`)}
          className="flex items-center justify-center gap-2 py-2 px-3 rounded-xl border border-black/10 bg-white hover:bg-zinc-50 text-xs font-semibold text-zinc-800 tactile-press cursor-pointer"
        >
          <Phone className="w-3.5 h-3.5 text-zinc-600" />
          <span>Call Driver</span>
        </button>
        <button
          type="button"
          onClick={() => alert(`Messaging driver ${driver.name}`)}
          className="flex items-center justify-center gap-2 py-2 px-3 rounded-xl border border-black/10 bg-white hover:bg-zinc-50 text-xs font-semibold text-zinc-800 tactile-press cursor-pointer"
        >
          <MessageSquare className="w-3.5 h-3.5 text-zinc-600" />
          <span>Send Message</span>
        </button>
      </div>
    </div>
  );
};
