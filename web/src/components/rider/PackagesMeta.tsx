import { CarPackageSlug } from "@/types";

export interface PackageMetaItem {
  name: string;
  icon: React.ReactNode;
  description: string;
  seats: number;
  badge?: string;
}

// Handcrafted Direction 3 SVG Vehicle Silhouettes (Zero generic placeholders)
export const SedanSvg = ({ className = "w-14 h-8" }: { className?: string }) => (
  <svg className={className} viewBox="0 0 100 48" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path
      d="M12 33 H88 C90 33 91 31 90 29 L86 21 C85 19 82 17 77 17 L62 17 L47 8 C44 6 36 6 30 11 L16 20 C13 22 10 25 10 28 L10 32 Z"
      fill="#18181B"
    />
    <circle cx="28" cy="33" r="6" fill="#FFFFFF" stroke="#18181B" strokeWidth="2.5" />
    <circle cx="74" cy="33" r="6" fill="#FFFFFF" stroke="#18181B" strokeWidth="2.5" />
    <path d="M34 18 L46 10 H59 L57 18 Z" fill="#FFFFFF" opacity="0.3" />
    <path d="M62 18 L64 10 H73 L79 18 Z" fill="#FFFFFF" opacity="0.3" />
  </svg>
);

export const SuvSvg = ({ className = "w-14 h-8" }: { className?: string }) => (
  <svg className={className} viewBox="0 0 100 48" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path
      d="M10 33 H90 C91.5 33 92 31 91 28 L86 16 C85 14 82 13 77 13 L38 13 C32 13 24 16 18 21 L11 27 C9 29 9 31 10 33 Z"
      fill="#3F3F46"
    />
    <circle cx="27" cy="33" r="6.5" fill="#FFFFFF" stroke="#3F3F46" strokeWidth="2.5" />
    <circle cx="75" cy="33" r="6.5" fill="#FFFFFF" stroke="#3F3F46" strokeWidth="2.5" />
    <rect x="36" y="15" width="22" height="7" rx="1.5" fill="#FFFFFF" opacity="0.3" />
    <rect x="62" y="15" width="16" height="7" rx="1.5" fill="#FFFFFF" opacity="0.3" />
    {/* Roof rails */}
    <line x1="38" y1="11" x2="78" y2="11" stroke="#18181B" strokeWidth="1.5" strokeLinecap="round" />
  </svg>
);

export const VanSvg = ({ className = "w-14 h-8" }: { className?: string }) => (
  <svg className={className} viewBox="0 0 100 48" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path
      d="M8 33 H92 C93 33 94 31 93 28 L88 12 C87 10 84 9 80 9 L24 9 C20 9 16 11 14 15 L9 24 C8 26 8 30 8 33 Z"
      fill="#27272A"
    />
    <circle cx="26" cy="33" r="6.5" fill="#FFFFFF" stroke="#27272A" strokeWidth="2.5" />
    <circle cx="76" cy="33" r="6.5" fill="#FFFFFF" stroke="#27272A" strokeWidth="2.5" />
    <rect x="24" y="12" width="18" height="9" rx="1.5" fill="#FFFFFF" opacity="0.25" />
    <rect x="46" y="12" width="20" height="9" rx="1.5" fill="#FFFFFF" opacity="0.25" />
    <rect x="70" y="12" width="12" height="9" rx="1.5" fill="#FFFFFF" opacity="0.25" />
  </svg>
);

export const LuxurySvg = ({ className = "w-14 h-8" }: { className?: string }) => (
  <svg className={className} viewBox="0 0 100 48" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path
      d="M12 33 H89 C91 33 92 31 91 29 L87 22 C85 19 81 17 76 17 L62 17 L44 9 C41 7 32 7 26 12 L14 20 C11 23 10 26 10 29 Z"
      fill="#09090B"
    />
    <circle cx="27" cy="33" r="6" fill="#FFFFFF" stroke="#D97706" strokeWidth="2" />
    <circle cx="75" cy="33" r="6" fill="#FFFFFF" stroke="#D97706" strokeWidth="2" />
    <line x1="20" y1="21" x2="84" y2="21" stroke="#D97706" strokeWidth="1" opacity="0.7" />
    <path d="M32 18 L44 11 H58 L56 18 Z" fill="#FFFFFF" opacity="0.35" />
    <path d="M60 18 L62 11 H71 L76 18 Z" fill="#FFFFFF" opacity="0.35" />
  </svg>
);

export const PackagesMeta: Record<CarPackageSlug, PackageMetaItem> = {
  [CarPackageSlug.SEDAN]: {
    name: "Classic Saloon",
    icon: <SedanSvg />,
    description: "Comfortable, prompt daily mobility",
    seats: 4,
    badge: "Popular",
  },
  [CarPackageSlug.SUV]: {
    name: "Tourer Estate",
    icon: <SuvSvg />,
    description: "Elevated room for luggage & company",
    seats: 6,
    badge: "Spacious",
  },
  [CarPackageSlug.VAN]: {
    name: "Executive Van",
    icon: <VanSvg />,
    description: "Multi-seat group travel with ease",
    seats: 8,
  },
  [CarPackageSlug.LUXURY]: {
    name: "First Class",
    icon: <LuxurySvg />,
    description: "Flagship quiet cabin & premier chauffeur",
    seats: 4,
    badge: "VIP",
  },
};