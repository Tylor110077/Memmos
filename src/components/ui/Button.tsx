import type { ButtonHTMLAttributes } from "react";
import { cn } from "@/lib/utils";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "ghost" | "link" | "danger";
  block?: boolean;
};

export function Button({ className, variant = "primary", block = false, ...props }: ButtonProps) {
  return (
    <button
      className={cn("btn", `btn-${variant}`, block && "btn-block", className)}
      {...props}
    />
  );
}
