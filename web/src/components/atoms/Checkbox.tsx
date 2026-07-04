interface Props {
  label: string
  checked: boolean
  disabled?: boolean
  onChange: (checked: boolean) => void
}

// フォーマット選択用のチップ型チェックボックス。選択状態はインジケータ
// ドットとアンバーの枠で示す。ラベル全体がタップ領域(FR-6)。
export function Checkbox({ label, checked, disabled, onChange }: Props) {
  const tone = checked
    ? 'border-amber/70 bg-amber/10 text-amber-bright'
    : 'border-line text-dim hover:border-line-strong hover:text-text'
  return (
    <label
      className={`flex min-h-11 cursor-pointer items-center gap-2.5 rounded-sm border px-3 font-mono text-xs tracking-[0.1em] uppercase transition-colors duration-150 ${tone} ${disabled ? 'pointer-events-none opacity-40' : ''}`}
    >
      <input
        type="checkbox"
        className="sr-only"
        checked={checked}
        disabled={disabled}
        onChange={(event) => onChange(event.target.checked)}
      />
      <span
        aria-hidden="true"
        className={`size-1.5 shrink-0 rounded-full transition-colors duration-150 ${checked ? 'bg-amber' : 'bg-faint/60'}`}
      />
      {label}
    </label>
  )
}
