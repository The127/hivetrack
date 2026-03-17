import { cva } from 'class-variance-authority'

const buttonVariants = cva(
  [
    'inline-flex items-center justify-center rounded-md font-medium',
    'transition-colors duration-150 cursor-pointer',
    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-1',
    'disabled:opacity-50 disabled:cursor-not-allowed disabled:pointer-events-none',
  ],
  {
    variants: {
      variant: {
        primary:     'bg-blue-600 text-white hover:bg-blue-700 active:bg-blue-800 focus-visible:ring-blue-500',
        secondary:   'bg-slate-100 text-slate-700 hover:bg-slate-200 active:bg-slate-300 focus-visible:ring-slate-400',
        ghost:       'text-slate-600 hover:bg-slate-100 hover:text-slate-900 active:bg-slate-200 focus-visible:ring-slate-400',
        destructive: 'bg-red-600 text-white hover:bg-red-700 active:bg-red-800 focus-visible:ring-red-500',
        link:        'text-blue-600 hover:underline focus-visible:ring-blue-500 h-auto px-0',
      },
      size: {
        sm: 'h-7 px-2.5 text-xs gap-1',
        md: 'h-8 px-3.5 text-sm gap-1.5',
        lg: 'h-9 px-4 text-sm gap-2',
      },
    },
    compoundVariants: [
      // link variant ignores size — reset padding/height applied by size variants
      { variant: 'link', size: 'sm', class: 'h-auto px-0' },
      { variant: 'link', size: 'md', class: 'h-auto px-0' },
      { variant: 'link', size: 'lg', class: 'h-auto px-0' },
    ],
    defaultVariants: {
      variant: 'primary',
      size: 'md',
    },
  },
)

export default buttonVariants
