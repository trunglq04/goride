'use client';

import React from 'react';
import { useRouter } from 'next/navigation';
import { Lock, ArrowRight, X, AlertTriangle, ShieldCheck } from 'lucide-react';

interface AuthRequiredModalProps {
  isOpen: boolean;
  onClose: () => void;
  role?: 'rider' | 'driver';
  errorMessage?: string | null;
}

export function AuthRequiredModal({
  isOpen,
  onClose,
  role = 'rider',
  errorMessage,
}: AuthRequiredModalProps) {
  const router = useRouter();

  if (!isOpen) return null;

  const isDriver = role === 'driver';

  return (
    <div className="fixed inset-0 z-[3000] bg-black/40 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-white border border-black/10 rounded-3xl p-6 sm:p-7 shadow-2xl flex flex-col items-center text-center relative animate-in fade-in zoom-in-95 duration-200">
        {/* Close Button */}
        <button
          type="button"
          onClick={onClose}
          className="absolute top-4 right-4 p-2 rounded-full text-zinc-400 hover:text-zinc-700 hover:bg-zinc-100 transition-colors cursor-pointer"
        >
          <X className="w-4 h-4" />
        </button>

        {/* Lock Icon Badge */}
        <div className="w-14 h-14 rounded-2xl bg-amber-50 border border-amber-200/80 flex items-center justify-center mb-4 text-amber-600 shadow-sm">
          <Lock className="w-7 h-7" />
        </div>

        {/* Title */}
        <h3 className="font-brand text-2xl sm:text-3xl font-extrabold text-[#18181b] tracking-tight">
          {isDriver ? 'Driver Sign In Required' : 'Account Sign In Required'}
        </h3>

        {/* Specific Backend Error Callout */}
        <div className="my-3 p-3 bg-amber-500/10 border border-amber-500/20 rounded-xl text-left w-full">
          <p className="text-xs text-amber-800/90 mt-1 leading-relaxed">
            {errorMessage ||
              (isDriver
                ? 'You must be authenticated with an authorized Driver Partner account to broadcast live vehicle GPS coordinates and receive dispatches.'
                : 'You must be signed in to your GoRide account to calculate real-time route fares and book a ride.')}
          </p>
        </div>

        {/* Trust Badges */}
        <div className="w-full py-2 flex items-center justify-center gap-4 text-[11px] text-zinc-500 border-y border-black/5 my-2">
          <span className="flex items-center gap-1">
            <ShieldCheck className="w-3.5 h-3.5 text-emerald-600" />
            <span>Encrypted Session</span>
          </span>
          <span>•</span>
          <span>Instant OTP Verification</span>
        </div>

        {/* Action Buttons */}
        <div className="w-full flex flex-col gap-2 mt-3">
          <button
            type="button"
            onClick={() => router.push('/login')}
            className="w-full py-3.5 px-4 rounded-xl font-semibold text-xs text-white bg-[#18181b] hover:bg-zinc-800 active:scale-[0.98] tactile-press shadow-sm cursor-pointer flex items-center justify-center gap-2"
          >
            <span>Sign In to GoRide</span>
            <ArrowRight className="w-3.5 h-3.5" />
          </button>

          <button
            type="button"
            onClick={() => router.push('/register')}
            className="w-full py-3 px-4 rounded-xl font-semibold text-xs text-zinc-700 bg-zinc-100 hover:bg-zinc-200 active:scale-[0.98] tactile-press cursor-pointer"
          >
            Create New Account
          </button>

          <button
            type="button"
            onClick={onClose}
            className="text-xs font-medium text-zinc-400 hover:text-zinc-600 py-1 transition-colors cursor-pointer mt-1"
          >
            Continue Browsing Map
          </button>
        </div>
      </div>
    </div>
  );
}
