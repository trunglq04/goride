"use client";

import React from "react";
import Link from "next/link";

interface BrandLogoProps {
  size?: "sm" | "md" | "lg";
  showText?: boolean;
  href?: string;
  className?: string;
  onClick?: () => void;
}

// Option 3: The Velocity Loop ("Continuous Highway Ribbon G" with emerald acceleration waypoint)
export const LogoMark = ({ className = "w-8 h-8" }: { className?: string }) => (
  <svg
    className={className}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <rect width="48" height="48" rx="14" fill="#09090B" />
    {/* Continuous highway ribbon forming 'G' */}
    <path
      d="M14 24C14 18.48 18.48 14 24 14C29.52 14 34 18.48 34 24C34 29.52 29.52 34 24 34H16"
      stroke="#FFFFFF"
      strokeWidth="4"
      strokeLinecap="round"
    />
    {/* Velocity route vector */}
    <path
      d="M24 24H34"
      stroke="#10B981"
      strokeWidth="4"
      strokeLinecap="round"
    />
    <circle cx="16" cy="34" r="2.5" fill="#10B981" />
  </svg>
);

export function BrandLogo({
  size = "md",
  showText = true,
  href,
  className = "",
  onClick,
}: BrandLogoProps) {
  const sizeClasses = {
    sm: {
      mark: "w-7 h-7",
      text: "text-lg",
    },
    md: {
      mark: "w-8 h-8",
      text: "text-2xl",
    },
    lg: {
      mark: "w-10 h-10",
      text: "text-3xl",
    },
  }[size];

  const content = (
    <div
      className={`inline-flex items-center gap-2.5 select-none cursor-pointer group ${className}`}
      onClick={onClick}
    >
      <div className="shrink-0 transition-transform duration-200 group-hover:scale-105 shadow-sm">
        <LogoMark className={sizeClasses.mark} />
      </div>

      {showText && (
        <span
          className={`font-brand font-extrabold text-[#18181b] tracking-tight ${sizeClasses.text} flex items-center`}
        >
          <span>Go</span>
          <span className="text-emerald-600">Ride</span>
        </span>
      )}
    </div>
  );

  if (href) {
    return (
      <Link href={href} className="inline-flex">
        {content}
      </Link>
    );
  }

  return content;
}
