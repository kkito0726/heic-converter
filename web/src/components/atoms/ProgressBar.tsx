interface Props {
  value: number
  max: number
}

export function ProgressBar({ value, max }: Props) {
  const percent = max === 0 ? 0 : Math.round((value / max) * 100)
  return (
    <div
      role="progressbar"
      aria-valuenow={value}
      aria-valuemin={0}
      aria-valuemax={max}
      className="h-2 w-full overflow-hidden rounded-full bg-slate-200"
    >
      <div className="h-full bg-indigo-600 transition-all" style={{ width: `${percent}%` }} />
    </div>
  )
}
