// 処理中を示す回転インジケータ。
export function Spinner() {
  return (
    <span
      role="status"
      aria-label="loading"
      className="inline-block size-4 animate-spin rounded-full border-2 border-current border-t-transparent text-indigo-600"
    />
  )
}
