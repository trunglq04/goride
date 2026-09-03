'use client';

// Assets
import 'leaflet/dist/leaflet.css';
import dynamic from 'next/dynamic';
import { Button } from '@/components/ui/button';
import { useState, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { CarPackageSlug } from '@/types';
import { DriverPackageSelector } from '@/components/driver/DriverPackageSelector';
import { useAuth } from '@/contexts/AuthContext';
import { SedanSvg, SuvSvg } from '@/components/rider/PackagesMeta';
import { ArrowRight, CheckCircle2, ShieldCheck, User, Compass, Car } from 'lucide-react';
import { BrandLogo } from '@/components/common/BrandLogo';

// Dynamically import components that use Leaflet
const DriverMap = dynamic(
  () => import('@/components/driver/DriverMap').then((mod) => mod.DriverMap),
  { ssr: false }
);
const RiderMap = dynamic(() => import('@/components/rider/RiderMap'), {
  ssr: false,
});

function HomeContent() {
  const [userType, setUserType] = useState<'driver' | 'rider' | null>(null);
  const router = useRouter();
  const searchParams = useSearchParams();
  const payment = searchParams.get('payment');
  const [packageSlug, setPackageSlug] = useState<CarPackageSlug | null>(null);

  const { user, isAuthenticated, logout } = useAuth();

  const handleClick = (selectedRole: 'driver' | 'rider') => {
    setUserType(selectedRole);
  };

  if (payment === 'success') {
    return (
      <main className="min-h-screen bg-[#f4f4f6] text-[#18181b] flex items-center justify-center p-4">
        <div className="w-full max-w-md bg-white border border-black/10 rounded-3xl p-8 text-center shadow-lg flex flex-col items-center">
          <div className="w-16 h-16 bg-emerald-50 border border-emerald-200/80 rounded-full flex items-center justify-center mb-5 text-emerald-600">
            <CheckCircle2 className="w-9 h-9" />
          </div>
          <h1 className="font-serif-brand text-3xl font-bold text-[#18181b] tracking-tight">
            Journey Confirmed
          </h1>
          <p className="text-xs text-zinc-500 mt-2.5 mb-6 leading-relaxed max-w-xs">
            Your ride payment has been authorized through Stripe Escrow. Your driver receipt is secured.
          </p>

          <Button
            className="w-full py-6 rounded-2xl bg-[#18181b] hover:bg-zinc-800 text-white font-semibold text-xs tactile-press cursor-pointer"
            onClick={() => router.push('/')}
          >
            Return to Transit Console
          </Button>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[#f4f4f6] text-[#18181b] flex flex-col">
      {/* 1. Direction 3 Top Navigation Bar */}
      <header className="sticky top-0 z-50 w-full border-b border-black/[0.08] bg-[#ffffff]/90 backdrop-blur-xl">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between">
          {/* Logo & Brand (Velocity Loop Option 3) */}
          <BrandLogo
            size="md"
            onClick={() => {
              setUserType(null);
              setPackageSlug(null);
            }}
          />

          {/* Role Indicator & Actions */}
          <div className="flex items-center gap-3">
            {userType && (
              <div className="hidden sm:flex items-center gap-1.5 px-3 py-1 rounded-full bg-zinc-100 border border-black/5 text-xs font-semibold text-zinc-700">
                <span className="w-2 h-2 rounded-full bg-emerald-600 animate-pulse" />
                <span className="capitalize">{userType} Mode</span>
                <button
                  type="button"
                  onClick={() => setUserType(userType === 'rider' ? 'driver' : 'rider')}
                  className="ml-1 text-[11px] text-zinc-500 hover:text-zinc-900 underline underline-offset-2 cursor-pointer"
                >
                  Switch
                </button>
              </div>
            )}

            {isAuthenticated && user ? (
              <div className="flex items-center gap-2.5">
                <div className="hidden sm:flex flex-col text-right">
                  <span className="text-xs font-bold text-[#18181b]">
                    {user.fullName}
                  </span>
                  <span className="text-[10px] font-mono text-zinc-500 uppercase tracking-wider">
                    {user.role}
                  </span>
                </div>
                <div className="w-8 h-8 rounded-full bg-zinc-100 border border-black/10 flex items-center justify-center text-xs font-bold text-zinc-800">
                  {user.fullName.slice(0, 2).toUpperCase()}
                </div>
                <button
                  type="button"
                  onClick={() => logout()}
                  className="px-3 py-1.5 rounded-xl border border-black/10 bg-white hover:bg-zinc-50 text-xs font-medium text-zinc-700 transition-colors cursor-pointer"
                >
                  Sign Out
                </button>
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => router.push('/login')}
                  className="px-3.5 py-1.5 rounded-xl text-xs font-semibold text-zinc-700 hover:text-[#18181b] transition-colors cursor-pointer"
                >
                  Sign In
                </button>
                <button
                  type="button"
                  onClick={() => router.push('/register')}
                  className="px-4 py-2 rounded-xl text-xs font-semibold text-white bg-[#18181b] hover:bg-zinc-800 shadow-sm transition-all tactile-press cursor-pointer"
                >
                  Get Started
                </button>
              </div>
            )}
          </div>
        </div>
      </header>

      {/* 2. Main Body Content Area */}
      <div className="flex-1 flex flex-col">
        {/* Landing Role Selection Modal */}
        {userType === null && (
          <div className="flex-1 flex flex-col items-center justify-center px-4 py-12">
            <div className="w-full max-w-xl bg-white border border-black/10 rounded-3xl p-8 sm:p-10 shadow-lg text-center flex flex-col items-center">
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-zinc-100 border border-black/5 text-zinc-700 text-xs font-medium mb-4">
                <ShieldCheck className="w-3.5 h-3.5 text-emerald-600" />
                <span>Next-Generation Urban Mobility</span>
              </div>

              <h2 className="font-serif-brand text-3xl sm:text-4xl font-bold text-[#18181b] tracking-tight max-w-md">
                Where would you like to travel today?
              </h2>
              <p className="text-xs sm:text-sm text-zinc-500 mt-2.5 mb-8 leading-relaxed max-w-sm mx-auto">
                Real-time dispatch, guaranteed transparent pricing, and instant live driver tracking.
              </p>

              {/* Two Role Options (Direction 3 Tactile Cards) */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 w-full mb-6">
                {/* Option 1: Rider */}
                <button
                  type="button"
                  onClick={() => handleClick('rider')}
                  className="p-5 rounded-2xl bg-[#fafafb] hover:bg-white border border-black/10 hover:border-[#18181b] hover:shadow-md transition-all text-left flex flex-col justify-between group cursor-pointer tactile-press"
                >
                  <div>
                    <div className="p-2 mb-3 inline-block">
                      <SedanSvg className="w-16 h-9" />
                    </div>
                    <div className="text-base font-bold text-[#18181b] group-hover:text-black">
                      I Need a Ride
                    </div>
                    <p className="text-xs text-zinc-500 mt-1">
                      Choose vehicle class, view upfront fare, and track your chauffeur.
                    </p>
                  </div>
                  <div className="mt-4 flex items-center gap-1.5 text-xs font-semibold text-[#18181b]">
                    <span>Enter Rider Console</span>
                    <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-0.5 transition-transform" />
                  </div>
                </button>

                {/* Option 2: Driver */}
                <button
                  type="button"
                  onClick={() => handleClick('driver')}
                  className="p-5 rounded-2xl bg-[#fafafb] hover:bg-white border border-black/10 hover:border-[#18181b] hover:shadow-md transition-all text-left flex flex-col justify-between group cursor-pointer tactile-press"
                >
                  <div>
                    <div className="p-2 mb-3 inline-block">
                      <SuvSvg className="w-16 h-9" />
                    </div>
                    <div className="text-base font-bold text-[#18181b] group-hover:text-black">
                      I Want to Drive
                    </div>
                    <p className="text-xs text-zinc-500 mt-1">
                      Broadcast vehicle location, receive dispatches, and manage trip requests.
                    </p>
                  </div>
                  <div className="mt-4 flex items-center gap-1.5 text-xs font-semibold text-[#18181b]">
                    <span>Enter Driver Console</span>
                    <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-0.5 transition-transform" />
                  </div>
                </button>
              </div>

              {!isAuthenticated && (
                <div className="pt-4 border-t border-black/5 text-xs text-zinc-500">
                  Already registered?{' '}
                  <button
                    type="button"
                    onClick={() => router.push('/login')}
                    className="font-bold text-[#18181b] hover:underline cursor-pointer"
                  >
                    Sign In to Account
                  </button>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Driver Map Console */}
        {userType === 'driver' && packageSlug && (
          <DriverMap packageSlug={packageSlug} />
        )}

        {/* Driver Package Selector */}
        {userType === 'driver' && !packageSlug && (
          <DriverPackageSelector onSelect={setPackageSlug} />
        )}

        {/* Rider Console (Direction 3 Ambient Map + Bottom Drawer) */}
        {userType === 'rider' && <RiderMap />}
      </div>
    </main>
  );
}

export default function Home() {
  return (
    <Suspense
      fallback={
        <main className="min-h-screen bg-[#f4f4f6] flex items-center justify-center">
          <div className="bg-white p-8 rounded-3xl border border-black/10 shadow-sm text-center max-w-sm w-full">
            <div className="w-8 h-8 border-2 border-black/20 border-t-[#18181b] rounded-full animate-spin mx-auto mb-4" />
            <p className="font-serif-brand text-lg font-semibold text-[#18181b]">Loading GoRide...</p>
          </div>
        </main>
      }
    >
      <HomeContent />
    </Suspense>
  );
}
