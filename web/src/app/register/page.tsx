'use client';

import React, { useState, useRef, useEffect } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { AuthLayout } from '@/components/auth/AuthLayout';
import { useAuth } from '@/contexts/AuthContext';
import {
  User,
  Mail,
  Phone,
  Lock,
  Eye,
  EyeOff,
  ArrowRight,
  ArrowLeft,
  Loader2,
  AlertCircle,
  CheckCircle2,
  RefreshCw,
  MailCheck,
} from 'lucide-react';
import { SedanSvg, SuvSvg } from '@/components/rider/PackagesMeta';

export default function RegisterPage() {
  const router = useRouter();
  const { register, verifyOTP, resendOTP } = useAuth();

  // Stage: "form" | "otp"
  const [stage, setStage] = useState<'form' | 'otp'>('form');

  // Form inputs
  const [role, setRole] = useState<'rider' | 'driver'>('rider');
  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('+84');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  // OTP inputs (6 separate boxes)
  const [otpDigits, setOtpDigits] = useState<string[]>([
    '',
    '',
    '',
    '',
    '',
    '',
  ]);
  const [userId, setUserId] = useState<string>('');
  const [countdown, setCountdown] = useState<number>(600); // 10 minutes
  const [resendCooldown, setResendCooldown] = useState<number>(0);

  // Status states
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isSuccess, setIsSuccess] = useState(false);

  // Input refs for 6 OTP boxes
  const otpInputRefs = useRef<(HTMLInputElement | null)[]>([]);

  // Countdown timer effect during OTP stage
  useEffect(() => {
    if (stage !== 'otp') return;

    const timer = setInterval(() => {
      setCountdown((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);

    return () => clearInterval(timer);
  }, [stage]);

  // Resend cooldown timer
  useEffect(() => {
    if (resendCooldown <= 0) return;
    const timer = setInterval(() => {
      setResendCooldown((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);
    return () => clearInterval(timer);
  }, [resendCooldown]);

  // Format seconds to mm:ss
  const formatTimer = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  // Password strength calculation
  const getPasswordStrength = () => {
    let score = 0;
    if (password.length >= 8) score++;
    if (/[A-Z]/.test(password)) score++;
    if (/[0-9]/.test(password)) score++;
    if (/[^A-Za-z0-9]/.test(password)) score++;
    return score;
  };
  const strengthScore = getPasswordStrength();

  // Handle registration submission
  const handleRegisterSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // Validate phone number E.164 pattern
    const phoneRegex = /^\+[1-9][0-9]{1,14}$/;
    if (!phoneRegex.test(phone.trim())) {
      setError(
        'Phone number must follow international E.164 format (e.g. +84965267199).'
      );
      return;
    }

    if (password.length < 8) {
      setError('Password must contain at least 8 characters.');
      return;
    }

    setIsLoading(true);

    try {
      const res = await register({
        fullName: fullName.trim(),
        email: email.trim().toLowerCase(),
        phone: phone.trim(),
        password,
        role,
      });

      if (!res.success || !res.userId) {
        setError(
          res.error || 'Failed to create account. Please verify your details.'
        );
        setIsLoading(false);
        return;
      }

      setUserId(res.userId);
      setStage('otp');
      setResendCooldown(30);
      setIsLoading(false);
      setError(null);

      // Focus first OTP input on transition
      setTimeout(() => {
        otpInputRefs.current[0]?.focus();
      }, 200);
    } catch {
      setError('Network error. Unable to reach authentication server.');
      setIsLoading(false);
    }
  };

  // Handle OTP digit changes
  const handleOtpChange = (index: number, value: string) => {
    const char = value.slice(-1);
    if (char && !/^[0-9]$/.test(char)) return;

    const newDigits = [...otpDigits];
    newDigits[index] = char;
    setOtpDigits(newDigits);

    if (char && index < 5) {
      otpInputRefs.current[index + 1]?.focus();
    }
  };

  // Handle backspace navigation in OTP
  const handleOtpKeyDown = (
    index: number,
    e: React.KeyboardEvent<HTMLInputElement>
  ) => {
    if (e.key === 'Backspace' && !otpDigits[index] && index > 0) {
      otpInputRefs.current[index - 1]?.focus();
    }
  };

  // Handle paste full 6-digit code
  const handleOtpPaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault();
    const pasted = e.clipboardData.getData('text').trim();
    if (/^[0-9]{6}$/.test(pasted)) {
      const digits = pasted.split('');
      setOtpDigits(digits);
      otpInputRefs.current[5]?.focus();
    }
  };

  // Handle OTP Verification submission
  const handleVerifySubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const fullCode = otpDigits.join('');
    if (fullCode.length !== 6) {
      setError('Please input the complete 6-digit code.');
      return;
    }

    setError(null);
    setIsLoading(true);

    try {
      const res = await verifyOTP(userId, fullCode);
      if (!res.success) {
        setError(res.error || 'Invalid or expired authorization code.');
        setIsLoading(false);
        return;
      }

      setIsSuccess(true);
      setTimeout(() => {
        router.push('/');
      }, 800);
    } catch {
      setError('Failed to verify code. Please try again.');
      setIsLoading(false);
    }
  };

  // Handle resend OTP
  const handleResend = async () => {
    if (resendCooldown > 0 || isLoading) return;
    setError(null);
    setIsLoading(true);

    try {
      const res = await resendOTP(userId);
      if (!res.success) {
        setError(res.error || 'Failed to resend verification code.');
      } else {
        setResendCooldown(30);
        setOtpDigits(['', '', '', '', '', '']);
        otpInputRefs.current[0]?.focus();
      }
    } catch {
      setError('Failed to resend code.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthLayout
      title={stage === 'form' ? 'Create your account' : 'Verify email'}
      subtitle={
        stage === 'form'
          ? 'Join riders and chauffeurs on the GoRide transit network.'
          : `We've dispatched a 6-digit verification code to ${email}.`
      }
    >
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
          <span>Email confirmed. Redirecting to dashboard...</span>
        </div>
      )}

      {stage === 'form' ? (
        /* ================= STAGE 1: REGISTRATION FORM ================= */
        <form onSubmit={handleRegisterSubmit} className="space-y-4">
          {/* Direction 3 Tactile Role Selection Cards */}
          <div>
            <label className="block text-xs font-semibold text-zinc-700 mb-2">
              Select Account Role
            </label>
            <div className="grid grid-cols-2 gap-3">
              {/* Rider Card */}
              <button
                type="button"
                onClick={() => setRole('rider')}
                className={`p-3.5 rounded-2xl border text-left transition-all cursor-pointer tactile-press flex flex-col justify-between ${
                  role === 'rider'
                    ? 'bg-white border-[#18181b] shadow-sm ring-1 ring-[#18181b]'
                    : 'bg-[#fafafb] border-black/8 hover:border-black/20'
                }`}
              >
                <div className="p-1 mb-2 inline-block">
                  <SedanSvg className="w-12 h-6" />
                </div>
                <div>
                  <div className="text-sm font-bold text-[#18181b]">Rider</div>
                  <p className="text-[11px] text-zinc-500 mt-0.5 leading-snug">
                    Book cars & upfront fares
                  </p>
                </div>
              </button>

              {/* Driver Card */}
              <button
                type="button"
                onClick={() => setRole('driver')}
                className={`p-3.5 rounded-2xl border text-left transition-all cursor-pointer tactile-press flex flex-col justify-between ${
                  role === 'driver'
                    ? 'bg-white border-[#18181b] shadow-sm ring-1 ring-[#18181b]'
                    : 'bg-[#fafafb] border-black/8 hover:border-black/20'
                }`}
              >
                <div className="p-1 mb-2 inline-block">
                  <SuvSvg className="w-12 h-6" />
                </div>
                <div>
                  <div className="text-sm font-bold text-[#18181b]">
                    Driver Partner
                  </div>
                  <p className="text-[11px] text-zinc-500 mt-0.5 leading-snug">
                    Broadcast GPS & drive
                  </p>
                </div>
              </button>
            </div>
          </div>

          {/* Full Name */}
          <div>
            <label className="block text-xs font-semibold text-zinc-700 mb-1.5">
              Full Legal Name
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-zinc-400">
                <User className="h-4 w-4" />
              </div>
              <input
                type="text"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                required
                placeholder="Alex Morgan"
                className="w-full pl-10 pr-4 py-2.5 rounded-xl bg-[#fafafb] border border-black/10 text-[#18181b] placeholder:text-zinc-400 text-sm focus:outline-none focus:border-[#18181b] focus:bg-white transition-all shadow-sm"
              />
            </div>
          </div>

          {/* Email Address */}
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
                placeholder="alex@example.com"
                className="w-full pl-10 pr-4 py-2.5 rounded-xl bg-[#fafafb] border border-black/10 text-[#18181b] placeholder:text-zinc-400 text-sm focus:outline-none focus:border-[#18181b] focus:bg-white transition-all shadow-sm"
              />
            </div>
          </div>

          {/* Phone Number */}
          <div>
            <div className="flex items-center justify-between mb-1.5">
              <label className="block text-xs font-semibold text-zinc-700">
                Mobile Number
              </label>
              <span className="text-[10px] font-mono text-zinc-400">
                E.164 standard
              </span>
            </div>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-zinc-400">
                <Phone className="h-4 w-4" />
              </div>
              <input
                type="tel"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                required
                placeholder="+14155552671"
                className="w-full pl-10 pr-4 py-2.5 rounded-xl bg-[#fafafb] border border-black/10 text-[#18181b] placeholder:text-zinc-400 text-sm focus:outline-none focus:border-[#18181b] focus:bg-white transition-all shadow-sm"
              />
            </div>
          </div>

          {/* Password */}
          <div>
            <label className="block text-xs font-semibold text-zinc-700 mb-1.5">
              Password
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-zinc-400">
                <Lock className="h-4 w-4" />
              </div>
              <input
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                placeholder="Minimum 8 characters"
                className="w-full pl-10 pr-10 py-2.5 rounded-xl bg-[#fafafb] border border-black/10 text-[#18181b] placeholder:text-zinc-400 text-sm focus:outline-none focus:border-[#18181b] focus:bg-white transition-all shadow-sm"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute inset-y-0 right-0 pr-3.5 flex items-center text-zinc-400 hover:text-zinc-700 transition-colors cursor-pointer"
              >
                {showPassword ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </button>
            </div>

            {/* Password Strength Meter */}
            {password && (
              <div className="mt-2 space-y-1.5">
                <div className="flex gap-1.5 h-1">
                  {[1, 2, 3, 4].map((step) => (
                    <div
                      key={step}
                      className={`flex-1 rounded-full transition-colors ${
                        step <= strengthScore
                          ? strengthScore === 1
                            ? 'bg-rose-500'
                            : strengthScore === 2
                              ? 'bg-amber-500'
                              : strengthScore === 3
                                ? 'bg-zinc-800'
                                : 'bg-emerald-600'
                          : 'bg-zinc-200'
                      }`}
                    />
                  ))}
                </div>
                <div className="flex items-center justify-between text-[10px] text-zinc-500">
                  <span>
                    {strengthScore <= 1 && 'Weak password'}
                    {strengthScore === 2 && 'Fair'}
                    {strengthScore === 3 && 'Good'}
                    {strengthScore === 4 && 'Strong password'}
                  </span>
                  <span>Must be ≥ 8 chars</span>
                </div>
              </div>
            )}
          </div>

          {/* Submit Button */}
          <button
            type="submit"
            disabled={isLoading}
            className="w-full mt-2 py-3.5 px-4 rounded-xl font-semibold text-sm text-white bg-[#18181b] hover:bg-zinc-800 active:scale-[0.98] transition-all shadow-sm flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed tactile-press cursor-pointer"
          >
            {isLoading ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>Creating Account...</span>
              </>
            ) : (
              <>
                <span>Continue to Verification</span>
                <ArrowRight className="h-4 w-4" />
              </>
            )}
          </button>

          {/* Link to Login */}
          <div className="pt-2 text-center text-xs text-zinc-500">
            Already have an account?{' '}
            <Link
              href="/login"
              className="font-bold text-[#18181b] hover:underline underline-offset-4 transition-colors"
            >
              Sign in here
            </Link>
          </div>
        </form>
      ) : (
        /* ================= STAGE 2: OTP VERIFICATION ================= */
        <form onSubmit={handleVerifySubmit} className="space-y-6">
          <div className="text-center py-2">
            <div className="h-14 w-14 rounded-full bg-zinc-100 border border-black/5 text-[#18181b] flex items-center justify-center mx-auto mb-3 shadow-sm">
              <MailCheck className="h-7 w-7 text-emerald-600" />
            </div>
            <p className="text-xs text-zinc-500">
              Check your inbox for the code. It expires in:
            </p>
            <div className="text-lg font-mono font-bold text-[#18181b] mt-1">
              {formatTimer(countdown)}
            </div>
          </div>

          {/* 6 Individual Digit Inputs (Direction 3 Tactile Boxes) */}
          <div>
            <label className="block text-xs font-semibold text-zinc-700 text-center mb-3">
              Enter 6-Digit Verification Code
            </label>
            <div
              className="flex justify-center gap-2 sm:gap-3"
              onPaste={handleOtpPaste}
            >
              {otpDigits.map((digit, idx) => (
                <input
                  key={idx}
                  ref={(el) => {
                    otpInputRefs.current[idx] = el;
                  }}
                  type="text"
                  inputMode="numeric"
                  maxLength={1}
                  value={digit}
                  onChange={(e) => handleOtpChange(idx, e.target.value)}
                  onKeyDown={(e) => handleOtpKeyDown(idx, e)}
                  className="w-11 h-13 sm:w-12 sm:h-14 text-center text-xl font-bold font-mono rounded-xl bg-[#fafafb] border border-black/15 text-[#18181b] focus:outline-none focus:border-[#18181b] focus:bg-white transition-all shadow-sm"
                />
              ))}
            </div>
          </div>

          {/* Verify Action Button */}
          <button
            type="submit"
            disabled={isLoading || isSuccess || otpDigits.join('').length !== 6}
            className="w-full py-3.5 px-4 rounded-xl font-semibold text-sm text-white bg-[#18181b] hover:bg-zinc-800 active:scale-[0.98] transition-all shadow-sm flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed tactile-press cursor-pointer"
          >
            {isLoading ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>Confirming Code...</span>
              </>
            ) : isSuccess ? (
              <>
                <CheckCircle2 className="h-4 w-4" />
                <span>Verified!</span>
              </>
            ) : (
              <>
                <span>Complete Registration</span>
                <ArrowRight className="h-4 w-4" />
              </>
            )}
          </button>

          {/* Resend & Back options */}
          <div className="flex items-center justify-between text-xs pt-2">
            <button
              type="button"
              onClick={() => setStage('form')}
              className="text-zinc-500 hover:text-[#18181b] flex items-center gap-1 transition-colors cursor-pointer"
            >
              <ArrowLeft className="h-3.5 w-3.5" />
              <span>Edit Details</span>
            </button>

            <button
              type="button"
              disabled={resendCooldown > 0 || isLoading}
              onClick={handleResend}
              className="text-[#18181b] font-bold hover:underline disabled:text-zinc-400 flex items-center gap-1.5 transition-colors cursor-pointer disabled:cursor-not-allowed"
            >
              <RefreshCw
                className={`h-3 w-3 ${isLoading ? 'animate-spin' : ''}`}
              />
              <span>
                {resendCooldown > 0
                  ? `Resend in ${resendCooldown}s`
                  : 'Resend Code'}
              </span>
            </button>
          </div>
        </form>
      )}
    </AuthLayout>
  );
}
