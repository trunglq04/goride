import React from "react";

interface TripOverviewCardProps {
  title: string;
  description: string;
  children?: React.ReactNode;
}

export const TripOverviewCard = ({
  title,
  description,
  children,
}: TripOverviewCardProps) => {
  return (
    <div className="w-full bg-white border border-black/8 rounded-3xl p-6 shadow-sm flex flex-col gap-4">
      <div className="space-y-1">
        <h2 className="font-brand text-2xl font-extrabold text-[#18181b] tracking-tight">
          {title}
        </h2>
        <p className="text-xs text-zinc-500 leading-relaxed">{description}</p>
      </div>
      {children && <div className="w-full">{children}</div>}
    </div>
  );
};
