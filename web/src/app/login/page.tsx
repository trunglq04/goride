"use client";

import React, { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { AuthLayout } from "@/components/auth/AuthLayout";
import { useAuth } from "@/contexts/AuthContext";
import { Mail, Lock, Eye, EyeOff, ArrowRight, Loader2, AlertCircle, CheckCircle2, Shield } from "lucide-react";
import { SedanSvg, SuvSvg } from "@/components/rider/PackagesMeta";

export default function LoginPage() {
  const router = useRouter();
  const { login } = useAuth();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(true);
  const [rolePreference, setRolePreference] = useState<"rider" | "driver">("rider");

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isSuccess, setIsSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email || !password) {
      setError("Please enter both email and password.");
      return;
    }

    setError(null);
    setIsLoading(true);

    try {
      const res = await login(email, password);
      if (!res.success) {
        setError(res.error || "Invalid credentials. Please verify your email and password.");
        setIsLoading(false);
        return;
      }

      setIsSuccess(true);
      setTimeout(() => {
        router.push("/");
      }, 700);
    } catch {
      setError("Unable to connect to authorization service. Please check your network.");
      setIsLoading(false);
    }
  };

  // Quick-fill helper for development & testing
  const handleQuickFill = (role: "rider" | "driver") => {
    setRolePreference(role);
    if (role === "rider") {
      setEmail("rider@goride.com");
      setPassword("Password123!");
    } else {
      setEmail("driver@goride.com");
      setPassword("Password123!");
    }
    setError(null);
  };

  return (
    <AuthLayout
      title="Welcome back"
      subtitle="Sign in to your GoRide account to access routes, dispatch, and trip receipts."
    >
      {/* Direction 3 Tactile Role Switcher */}
      <div className="flex p-1 bg-[#f0f0f2] border border-black/5 rounded-2xl mb-6">
        <button
          type="button"
          onClick={() => setRolePreference("rider")}
          className={`flex-1 py-2 text-xs font-semibold rounded-xl transition-all flex items-center justify-center gap-2 cursor-pointer ${
            rolePreference === "rider"
              ? "bg-white text-[#18181b] shadow-sm"
              : "text-zinc-500 hover:text-zinc-800"
          }`}
        >
          <div className="w-5 h-3 flex items-center justify-center">
            <SedanSvg className="w-5 h-3" />
          </div>
          <span>Rider Account</span>
        </button>
        <button
          type="button"
          onClick={() => setRolePreference("driver")}
          className={`flex-1 py-2 text-xs font-semibold rounded-xl transition-all flex items-center justify-center gap-2 cursor-pointer ${
            rolePreference === "driver"
              ? "bg-white text-[#18181b] shadow-sm"
              : "text-zinc-500 hover:text-zinc-800"
          }`}
        >
          <div className="w-5 h-3 flex items-center justify-center">
            <SuvSvg className="w-5 h-3" />
          </div>
          <span>Driver Partner</span>
        </button>
      </div>

      {/* Error Alert */}
      {error && (
        <div className="mb-5 p-3.5 rounded-2xl bg-rose-50 border border-rose-200/80 text-rose-800 text-xs flex items-start gap-2.5">
          <AlertCircle className="h-4 w-4 text-rose-600 shrink-0 mt-0.5" />
          <div className="flex-1 leading-relaxed">{error}</div>
        </div>
      )}

      {/* Success Alert */}
      {isSuccess && (
        <div className="mb-5 p-3.5 rounded-2xl bg-emerald-50 border border-emerald-200 text-emerald-800 text-xs flex items-center gap-2.5">
          <CheckCircle2 className="h-4 w-4 text-emerald-600 shrink-0" />
          <span>Authenticated successfully. Loading console...</span>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Email Field */}
        <div>
          <label className="block text-xs font-semibold text-zinc-700 mb-1.5">
            Email Address
          </label>
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-zinc-400">
              <Mail className="h-4 w-4" />
            </div>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              placeholder="name@example.com"
              className="w-full pl-10 pr-4 py-2.5 rounded-xl bg-[#fafafb] border border-black/10 text-[#18181b] placeholder:text-zinc-400 text-sm focus:outline-none focus:border-[#18181b] focus:bg-white transition-all shadow-sm"
            />
          </div>
        </div>

        {/* Password Field */}
        <div>
          <div className="flex items-center justify-between mb-1.5">
            <label className="block text-xs font-semibold text-zinc-700">
              Password
            </label>
            <a
              href="#"
              onClick={(e) => {
                e.preventDefault();
                alert("Password reset instructions will be sent to your registered email.");
              }}
              className="text-xs text-zinc-500 hover:text-[#18181b] transition-colors"
            >
              Forgot password?
            </a>
          </div>
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-zinc-400">
              <Lock className="h-4 w-4" />
            </div>
            <input
              type={showPassword ? "text" : "password"}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              placeholder="••••••••"
              className="w-full pl-10 pr-10 py-2.5 rounded-xl bg-[#fafafb] border border-black/10 text-[#18181b] placeholder:text-zinc-400 text-sm focus:outline-none focus:border-[#18181b] focus:bg-white transition-all shadow-sm"
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute inset-y-0 right-0 pr-3.5 flex items-center text-zinc-400 hover:text-zinc-700 transition-colors cursor-pointer"
            >
              {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
            </button>
          </div>
        </div>

        {/* Remember me Checkbox */}
        <div className="flex items-center justify-between pt-1">
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={rememberMe}
              onChange={(e) => setRememberMe(e.target.checked)}
              className="h-4 w-4 rounded border-zinc-300 text-[#18181b] focus:ring-[#18181b]"
            />
            <span className="text-xs text-zinc-600 font-medium">
              Keep me signed in
            </span>
          </label>
        </div>

        {/* Submit Button (Direction 3 Tactile Solid Ink) */}
        <button
          type="submit"
          disabled={isLoading || isSuccess}
          className="w-full mt-2 py-3.5 px-4 rounded-xl font-semibold text-sm text-white bg-[#18181b] hover:bg-zinc-800 active:scale-[0.98] transition-all shadow-sm flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed tactile-press cursor-pointer"
        >
          {isLoading ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              <span>Authenticating...</span>
            </>
          ) : isSuccess ? (
            <>
              <CheckCircle2 className="h-4 w-4 text-white" />
              <span>Welcome Back</span>
            </>
          ) : (
            <>
              <span>Sign In to GoRide</span>
              <ArrowRight className="h-4 w-4" />
            </>
          )}
        </button>
      </form>

      {/* Developer Sandbox Quick Fill (Tactile Style) */}
      <div className="mt-6 pt-5 border-t border-black/5">
        <div className="flex items-center justify-between mb-2.5">
          <span className="text-[11px] font-semibold text-zinc-400 uppercase tracking-wider flex items-center gap-1.5">
            <Shield className="h-3 w-3 text-emerald-600" />
            <span>Developer Sandbox Presets</span>
          </span>
        </div>
        <div className="grid grid-cols-2 gap-2">
          <button
            type="button"
            onClick={() => handleQuickFill("rider")}
            className="p-2.5 rounded-xl bg-[#fafafb] hover:bg-white border border-black/8 hover:border-black/20 text-left transition-all cursor-pointer tactile-press"
          >
            <div className="text-xs font-bold text-[#18181b]">Rider Demo</div>
            <div className="text-[11px] text-zinc-400 font-mono truncate">rider@goride.com</div>
          </button>
          <button
            type="button"
            onClick={() => handleQuickFill("driver")}
            className="p-2.5 rounded-xl bg-[#fafafb] hover:bg-white border border-black/8 hover:border-black/20 text-left transition-all cursor-pointer tactile-press"
          >
            <div className="text-xs font-bold text-[#18181b]">Driver Demo</div>
            <div className="text-[11px] text-zinc-400 font-mono truncate">driver@goride.com</div>
          </button>
        </div>
      </div>

      {/* Link to Register */}
      <div className="mt-6 text-center text-xs text-zinc-500">
        Don&apos;t have an account yet?{" "}
        <Link
          href="/register"
          className="font-bold text-[#18181b] hover:underline underline-offset-4 transition-colors"
        >
          Create account
        </Link>
      </div>
    </AuthLayout>
  );
}
