interface Props {
  value: number
  max: number
}

// 2pxのヘアラインプログレス。走査光(シマー)で進行中を示す。
export function ProgressBar({ value, max }: Props) {
  const percent = max === 0 ? 0 : Math.round((value / max) * 100)
  return (
    <div
      role="progressbar"
      aria-valuenow={value}
      aria-valuemin={0}
      aria-valuemax={max}
      className="h-0.5 w-full overflow-hidden bg-line"
    >
      <div
        className="shimmer relative h-full bg-amber transition-[width] duration-300"
        style={{ width: `${percent}%` }}
      />
    </div>
  )
}
