import type { ReactNode } from 'react'

interface Props {
  dropZone: ReactNode
  fileList: ReactNode
  settings: ReactNode
  results: ReactNode
}

// モバイル1カラム / デスクトップ2カラムの骨組み(FR-6)。
// スマホでは 選ぶ → 設定 → 変換 → 共有 が縦スクロールで自然に流れる順に並ぶ。
export function ConverterLayout({ dropZone, fileList, settings, results }: Props) {
  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <main className="mx-auto flex max-w-4xl flex-col gap-4 p-4 sm:p-6">
        <header>
          <h1 className="text-2xl font-bold tracking-tight">HEIC Converter</h1>
          <p className="mt-1 text-sm text-slate-500">
            Convert .heic / .heif photos to jpg, png and more — right from your browser.
          </p>
        </header>
        <div className="grid items-start gap-4 md:grid-cols-[3fr_2fr]">
          <div className="flex flex-col gap-4">
            {dropZone}
            {fileList}
          </div>
          <div className="flex flex-col gap-4">
            {settings}
            {results}
          </div>
        </div>
        <footer className="pt-2 pb-6 text-center text-xs text-slate-400">
          Images are never stored on the server. Converted files exist only in this browser tab
          and disappear on reload.
        </footer>
      </main>
    </div>
  )
}
