import { ReactNode } from "react";

interface TabProps {
  label: string;
  children: ReactNode;
  className?: string;
}

export function Tab({ label, children, className = "" }: TabProps) {
  return (
    <div
      className={`flex items-center gap-3 justify-around text-md border rounded-md p-1 font-semibold uppercase ${className}`}
    >
      <p>{label}</p>
      <p>{children}</p>
    </div>
  );
}
