import type { ButtonHTMLAttributes } from 'react'

type Variant = 'primary' | 'secondary' | 'ghost'

const STYLES: Record<Variant, string> = {
  primary:
    'bg-amber text-ink hover:bg-amber-bright active:translate-y-px disabled:bg-panel-raised disabled:text-faint',
  secondary:
    'border border-line-strong bg-transparent text-text hover:border-amber hover:text-amber-bright disabled:border-line disabled:text-faint',
  ghost: 'text-faint hover:bg-panel-raised hover:text-text',
}

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
}

// 計器のキーを思わせるモノスペース・角の締まったボタン。
// 最小44pxのタップターゲットを保証する(FR-6)。
export function Button({ variant = 'primary', className = '', type = 'button', ...rest }: Props) {
  return (
    <button
      type={type}
      className={`inline-flex min-h-11 cursor-pointer items-center justify-center gap-2 rounded-sm px-4 font-mono text-xs font-medium tracking-[0.08em] uppercase transition-all duration-150 disabled:cursor-not-allowed ${STYLES[variant]} ${className}`}
      {...rest}
    />
  )
}
