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
    <div className="min-h-screen">
      {/* 機材の銘板を模したトップストリップ */}
      <div className="border-b border-line">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-2 sm:px-6">
          <span className="font-mono text-[10px] tracking-[0.22em] text-faint uppercase">
            heic-converter
          </span>
          <span className="hidden font-mono text-[10px] tracking-[0.14em] text-faint uppercase sm:block">
            local processing · nothing stored
          </span>
          <span className="flex items-center gap-1.5 font-mono text-[10px] tracking-[0.14em] text-ok uppercase">
            <span aria-hidden="true" className="size-1.5 rounded-full bg-ok" />
            ready
          </span>
        </div>
      </div>

      <main className="mx-auto flex max-w-5xl flex-col gap-5 px-4 py-8 sm:px-6 sm:py-12">
        <header className="reveal">
          <h1 className="text-4xl font-bold tracking-tight text-text sm:text-5xl">
            HEIC Converter
          </h1>
          <p className="mt-3 max-w-xl text-sm leading-relaxed text-dim">
            Convert .heic / .heif photos to jpg, png and more — right from your browser.
          </p>
        </header>

        <div className="grid items-start gap-5 md:grid-cols-[3fr_2fr]">
          <div className="flex flex-col gap-5">
            <div className="reveal" style={{ animationDelay: '80ms' }}>
              {dropZone}
            </div>
            <div className="reveal" style={{ animationDelay: '140ms' }}>
              {fileList}
            </div>
          </div>
          <div className="flex flex-col gap-5">
            <div className="reveal" style={{ animationDelay: '200ms' }}>
              {settings}
            </div>
            {results}
          </div>
        </div>

        <footer
          className="reveal mt-6 border-t border-line pt-5 font-mono text-[11px] leading-relaxed text-faint"
          style={{ animationDelay: '260ms' }}
        >
          Images are never stored on the server. Converted files exist only in this browser tab
          and disappear on reload.
        </footer>
      </main>
    </div>
  )
}
