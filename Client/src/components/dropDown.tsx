import React, { ReactNode, useState } from 'react';
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
    <div className="relative inline-block" onMouseLeave={() => setOpen(false)}>
      {React.Children.map(children, (child) => {
        if (React.isValidElement(child)) {
          if (child.type === DropdownMenuTrigger) {
            return React.cloneElement(child as React.ReactElement<any>, { 
              onClick: () => setOpen(!open),
              isOpen: open 
            });
          }
          if (child.type === DropdownMenuContent && open) {
            return child;
          }
          if (child.type === DropdownMenuContent && !open) {
            return null;
          }
        }
        return child;
      })}
    </div>
  );
}

export function DropdownMenuTrigger({ children, className, onClick, isOpen }: DropdownMenuTriggerProps & { onClick?: () => void; isOpen?: boolean }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={className}
      aria-expanded={isOpen}
    >
      {children}
    </button>
  );
}

export function DropdownMenuContent({ children, align, className }: DropdownMenuContentProps) {
  const alignClass = align === 'end' ? 'right-0' : align === 'center' ? 'left-1/2 -translate-x-1/2' : 'left-0';
  return (
    <div className={`absolute z-50 mt-2 min-w-[8rem] rounded-md border bg-secondary/30 backdrop-blur-sm text-foreground border-border/50 p-1 shadow-lg ${alignClass} ${className || ''}`}>
      {children}
    </div>
  );
}

export function DropdownMenuItem({ children, onClick, className }: DropdownMenuItemProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`block w-full px-3 py-2 text-left text-sm text-foreground hover:bg-background/50 rounded-sm transition-colors ${className ?? ""}`}
    >
      {children}
    </button>
  )
}

export function DropdownMenuSeparator({}: DropdownMenuSeparatorProps) {
  return <div className="my-1 h-px bg-border" />;
}
