import { createClient, type Client } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { ConvertService } from '../gen/heic/v1/convert_pb'

// ConvertServiceの型付きクライアント。hooksへはこの型で注入する
// (テストではフェイクに差し替える)。
export type ConverterClient = Client<typeof ConvertService>

// baseUrl未指定時は同一オリジン(将来のgo:embed同梱を想定)。
// 開発時は VITE_API_URL でserveのアドレス(例: http://localhost:8080)を指定する。
export function createConverterClient(baseUrl?: string): ConverterClient {
  const url = baseUrl ?? import.meta.env.VITE_API_URL ?? window.location.origin
  return createClient(ConvertService, createConnectTransport({ baseUrl: url }))
}
