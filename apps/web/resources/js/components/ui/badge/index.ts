import type { VariantProps } from "class-variance-authority"
import { cva } from "class-variance-authority"

export { default as Badge } from "./Badge.vue"

// The design system gives Badge a 6px radius (the pill shape belongs to Tag,
// not Badge) and a tint-background / strong-foreground pair per status tone.
export const badgeVariants = cva(
  "inline-flex items-center justify-center rounded-md border px-2 py-0.5 text-xs font-semibold w-fit whitespace-nowrap shrink-0 [&>svg]:size-3 gap-1 [&>svg]:pointer-events-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive transition-[color,box-shadow] overflow-hidden",
  {
    variants: {
      variant: {
        default:
          "border-transparent bg-primary text-primary-foreground [a&]:hover:bg-primary-hover",
        secondary:
          "border-transparent bg-accent text-secondary-foreground [a&]:hover:bg-accent/80",
        destructive:
          "border-transparent bg-danger-tint text-danger-strong [a&]:hover:bg-danger-tint/80",
        success:
          "border-transparent bg-success-tint text-success-strong [a&]:hover:bg-success-tint/80",
        warning:
          "border-transparent bg-warning-tint text-warning-strong [a&]:hover:bg-warning-tint/80",
        info:
          "border-transparent bg-info-tint text-info-strong [a&]:hover:bg-info-tint/80",
        outline:
          "text-foreground [a&]:hover:bg-accent [a&]:hover:text-accent-foreground",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
)
export type BadgeVariants = VariantProps<typeof badgeVariants>
