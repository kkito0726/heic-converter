import { create } from '@bufbuild/protobuf'
import {
  ConvertResponseSchema,
  ListFormatsResponseSchema,
} from '../gen/heic/v1/convert_pb'
import type { ConverterClient } from '../lib/rpc'

export const FAKE_FORMATS = ['jpg', 'png', 'webp']

// 要求された全形式にダミーのバイト列を返すフェイククライアント。
// 個別のテストではメソッドを差し替えて失敗や遅延を注入する。
export function fakeClient(overrides: Partial<ConverterClient> = {}): ConverterClient {
  const base: ConverterClient = {
    convert: async (req) =>
      create(ConvertResponseSchema, {
        images: (req.formats ?? []).map((format) => ({
          format,
          data: new Uint8Array([1, 2, 3]),
        })),
      }),
    listFormats: async () => create(ListFormatsResponseSchema, { formats: FAKE_FORMATS }),
  }
  return { ...base, ...overrides }
}

// signalが中断されるまで解決しないconvert実装(キャンセルのテスト用)。
export function hangingConvert(): ConverterClient['convert'] {
  return (_req, options) =>
    new Promise((_resolve, reject) => {
      options?.signal?.addEventListener('abort', () =>
        reject(new DOMException('aborted', 'AbortError')),
      )
    })
}

// テスト用のHEICファイルを作る。
export function heicFile(name = 'photo.heic', content = 'fake-heic'): File {
  return new File([content], name, { type: 'image/heic' })
}
