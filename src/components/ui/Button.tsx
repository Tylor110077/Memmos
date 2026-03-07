import { forwardRef, type ButtonHTMLAttributes } from "react";
import { cn } from "@/lib/utils";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "ghost" | "link" | "danger";
  block?: boolean;
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { className, variant = "primary", block = false, ...props },
  ref,
) {
  return <button ref={ref} className={cn("btn", `btn-${variant}`, block && "btn-block", className)} {...props} />;
});
