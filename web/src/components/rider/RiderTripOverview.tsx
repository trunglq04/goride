import React from "react";
import { RouteFare, TripPreview, Driver } from "@/types";
import { DriverList } from "./DriversList";
import { Button } from "@/components/ui/button";
import {
  convertMetersToKilometers,
  convertSecondsToMinutes,
} from "@/utils/math";
import { TripOverviewCard } from "./TripOverviewCard";
import { StripePaymentButton } from "@/components/common/StripePaymentButton";
import { DriverCard } from "@/components/driver/DriverCard";
import { TripEvents, PaymentEventSessionCreatedData } from "@/types/contracts";
import { CheckCircle2, AlertCircle, XCircle, Search, ShieldCheck } from "lucide-react";

interface TripOverviewProps {
  trip: TripPreview | null;
  status: TripEvents | null;
  assignedDriver?: Driver | null;
  paymentSession?: PaymentEventSessionCreatedData | null;
  onPackageSelect: (carPackage: RouteFare) => void;
  onCancel: () => void;
}

export const RiderTripOverview = ({
  trip,
  status,
  assignedDriver,
  paymentSession,
  onPackageSelect,
  onCancel,
}: TripOverviewProps) => {
  // State 1: Initial state before route selection
  if (!trip) {
    return (
      <TripOverviewCard
        title="Where to?"
        description="Tap anywhere on the map or choose a popular destination to preview routes and fares."
      >
        <div className="flex items-center gap-3 p-3.5 bg-[#fafafb] border border-black/5 rounded-2xl text-xs text-zinc-500">
          <div className="w-2.5 h-2.5 rounded-full bg-[#18181b]" />
          <span>Click anywhere on the map to set your destination pin.</span>
        </div>
      </TripOverviewCard>
    );
  }

  // State 2: Payment Required (Stripe Escrow)
  if (status === TripEvents.PaymentSessionCreated && paymentSession) {
    const formattedAmount = (paymentSession.amount / 100).toFixed(2);

    return (
      <TripOverviewCard
        title="Confirm & Authorize"
        description="Authorize payment through Stripe escrow. Funds are securely captured upon arrival."
      >
        <div className="flex flex-col gap-4">
          <DriverCard driver={assignedDriver} />

          {/* Receipt Breakdown Card */}
          <div className="p-4 bg-[#fafafb] border border-black/5 rounded-2xl flex flex-col gap-2.5 text-xs">
            <div className="flex justify-between items-center text-zinc-500">
              <span>Trip Ledger Reference</span>
              <span className="font-mono text-zinc-800">
                #{paymentSession.tripID.slice(0, 8)}
              </span>
            </div>
            <div className="flex justify-between items-center text-zinc-500">
              <span>Currency & Escrow Type</span>
              <span className="uppercase text-zinc-800 font-semibold">
                {paymentSession.currency} (Stripe Verified)
              </span>
            </div>
            <div className="h-[1px] bg-black/5 my-1" />
            <div className="flex justify-between items-center">
              <span className="font-bold text-sm text-[#18181b]">Total Authorized</span>
              <span className="font-brand font-extrabold text-xl text-[#18181b]">
                ${formattedAmount}
              </span>
            </div>
          </div>

          <div className="flex items-center gap-2 text-[11px] text-zinc-500">
            <ShieldCheck className="w-4 h-4 text-emerald-600 shrink-0" />
            <span>Encrypted Stripe checkout with instant ride insurance.</span>
          </div>

          <StripePaymentButton paymentSession={paymentSession} />
        </div>
      </TripOverviewCard>
    );
  }

  // State 3: No Drivers Found
  if (status === TripEvents.NoDriversFound) {
    return (
      <TripOverviewCard
        title="No Drivers Nearby"
        description="All drivers in your immediate dispatch radius are currently on trips. Please try again in a moment."
      >
        <div className="flex flex-col gap-3">
          <div className="p-4 bg-amber-50 border border-amber-200/60 rounded-2xl flex items-center gap-3 text-xs text-amber-800">
            <AlertCircle className="w-5 h-5 text-amber-600 shrink-0" />
            <span>You can expand your route or check back shortly as trips complete.</span>
          </div>
          <Button
            variant="outline"
            className="w-full py-5 rounded-2xl text-xs font-semibold cursor-pointer"
            onClick={onCancel}
          >
            Return to Map
          </Button>
        </div>
      </TripOverviewCard>
    );
  }

  // State 4: Driver Assigned & En Route
  if (status === TripEvents.DriverAssigned) {
    return (
      <TripOverviewCard
        title="Driver is on the way!"
        description="Your vehicle has been dispatched. Preparing secure payment authorization..."
      >
        <div className="flex flex-col gap-4">
          <DriverCard driver={assignedDriver} />

          <Button
            variant="destructive"
            className="w-full py-5 rounded-2xl text-xs font-semibold tactile-press cursor-pointer"
            onClick={onCancel}
          >
            Cancel Trip Request
          </Button>
        </div>
      </TripOverviewCard>
    );
  }

  // State 5: Completed
  if (status === TripEvents.Completed) {
    return (
      <TripOverviewCard
        title="Journey Completed"
        description="Thank you for traveling with GoRide. Your electronic receipt has been dispatched to your email."
      >
        <div className="flex flex-col gap-4 items-center text-center py-4">
          <div className="w-14 h-14 rounded-full bg-emerald-50 border border-emerald-200 flex items-center justify-center text-emerald-600">
            <CheckCircle2 className="w-8 h-8" />
          </div>
          <p className="text-xs text-zinc-500 max-w-xs">
            We hope you enjoyed the ride. Please rate your driver to help keep our community five-star.
          </p>
          <Button
            className="w-full py-5 rounded-2xl bg-[#18181b] text-white text-xs font-semibold tactile-press cursor-pointer"
            onClick={onCancel}
          >
            Book Another Journey
          </Button>
        </div>
      </TripOverviewCard>
    );
  }

  // State 6: Cancelled
  if (status === TripEvents.Cancelled) {
    return (
      <TripOverviewCard
        title="Trip Cancelled"
        description="This journey request has been cancelled. No cancellation penalty was charged."
      >
        <div className="flex flex-col gap-3">
          <div className="p-4 bg-zinc-50 border border-black/5 rounded-2xl flex items-center gap-3 text-xs text-zinc-600">
            <XCircle className="w-5 h-5 text-zinc-400 shrink-0" />
            <span>Your dispatch hold has been immediately released.</span>
          </div>
          <Button
            variant="outline"
            className="w-full py-5 rounded-2xl text-xs font-semibold cursor-pointer"
            onClick={onCancel}
          >
            Return to Map
          </Button>
        </div>
      </TripOverviewCard>
    );
  }

  // State 7: Created (Matching / Looking for a Driver)
  if (status === TripEvents.Created) {
    return (
      <TripOverviewCard
        title="Matching your Driver"
        description="Connecting with nearby vehicles on the real-time dispatch stream..."
      >
        <div className="flex flex-col items-center justify-center py-6 gap-4">
          {/* Animated Radar Pulse */}
          <div className="relative flex items-center justify-center">
            <div className="w-20 h-20 rounded-full bg-emerald-500/10 animate-ping absolute" />
            <div className="w-16 h-16 rounded-full bg-emerald-500/20 border border-emerald-500/30 flex items-center justify-center z-10">
              <Search className="w-6 h-6 text-emerald-700 animate-pulse" />
            </div>
          </div>

          <div className="text-center space-y-1">
            <p className="text-xs font-bold text-[#18181b]">Broadcasting dispatch request</p>
            <p className="text-[11px] text-zinc-500">
              ETA to destination: ~{convertSecondsToMinutes(trip?.duration ?? 0)} ({convertMetersToKilometers(trip?.distance ?? 0)})
            </p>
          </div>

          <Button
            variant="outline"
            className="w-full py-5 rounded-2xl text-xs font-semibold border-black/10 text-zinc-600 hover:text-zinc-900 cursor-pointer"
            onClick={onCancel}
          >
            Cancel Request
          </Button>
        </div>
      </TripOverviewCard>
    );
  }

  // State 8: Ride Selection Phase
  if (trip.rideFares && trip.rideFares.length >= 0 && !trip.tripID) {
    return (
      <TripOverviewCard
        title="Select Vehicle"
        description="Compare vehicle classes, luggage capacities, and guaranteed route pricing."
      >
        <DriverList
          trip={trip}
          onPackageSelect={onPackageSelect}
          onCancel={onCancel}
        />
      </TripOverviewCard>
    );
  }

  return null;
};
