import { ReactNode, useState } from 'react';
import type { MouseEventHandler } from "react"

interface DropdownMenuProps {
  children: ReactNode;
}

interface DropdownMenuTriggerProps {
  children: ReactNode;
  className?: string;
}

interface DropdownMenuContentProps {
  children: ReactNode;
  align?: 'start' | 'end' | 'center';
  className?: string;
}

interface DropdownMenuItemProps {
  children: ReactNode;
onClick?: MouseEventHandler<HTMLButtonElement>;
  className?: string;
}

interface DropdownMenuSeparatorProps {}

export function DropdownMenu({ children }: DropdownMenuProps) {
  const [open, setOpen] = useState(false);
  
  return (
    <div className="relative inline-block">
      {children}
    </div>
  );
}

export function DropdownMenuTrigger({ children, className }: DropdownMenuTriggerProps) {
  return <div className={className}>{children}</div>;
}

export function DropdownMenuContent({ children, align, className }: DropdownMenuContentProps) {
  return (
    <div className={`absolute z-50 mt-2 min-w-[8rem] rounded-md border bg-popover p-1 shadow-md ${className || ''}`}>
      {children}
    </div>
  );
}

export function DropdownMenuItem({ children, onClick, className }: DropdownMenuItemProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`block w-full px-3 py-2 text-left text-sm hover:bg-gray-100 ${className ?? ""}`}
    >
      {children}
    </button>
  )
}

export function DropdownMenuSeparator({}: DropdownMenuSeparatorProps) {
  return <div className="my-1 h-px bg-border" />;
}
