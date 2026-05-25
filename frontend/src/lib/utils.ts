import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

// cn() merges Tailwind classes safely — used in every shadcn component.
// e.g. cn('px-4', condition && 'bg-red-500', props.class)
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
