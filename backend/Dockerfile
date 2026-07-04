# APIサーバー(heic-converter serve)のイメージ。
FROM golang:1.26 AS build
WORKDIR /src

# 依存ダウンロードを別レイヤーにしてビルドキャッシュを効かせる
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /heic-converter ./cmd/heic-converter

# cgo不使用だが、依存のpurego(gen2brain系のFFI層)が動的ローダー(ld-linux)を
# 要求するためscratchでは動かない。glibcだけ入った最小・シェルなし・非rootの
# distroless/baseを使う。
FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=build /heic-converter /heic-converter
EXPOSE 8080
ENTRYPOINT ["/heic-converter", "serve"]
