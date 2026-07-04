package model

// DefaultQualityは品質が指定されなかった場合(または不正な値の場合)に
// 使われるエンコード品質。
const DefaultQuality = 90

// EncodeOptionsは全エンコーダで共有されるエンコード調整パラメータを保持する。
type EncodeOptions struct {
	// Qualityは1-100の範囲の非可逆圧縮品質。
	// 可逆形式ではこの値は無視される。
	Quality int
}

// NewEncodeOptionsはEncodeOptionsを構築する。qualityが1-100の範囲外のときは
// DefaultQualityにフォールバックする。
func NewEncodeOptions(quality int) EncodeOptions {
	if quality < 1 || quality > 100 {
		quality = DefaultQuality
	}
	return EncodeOptions{Quality: quality}
}

// ConversionResultは1つの変換元ファイルを変換した結果を表す。
type ConversionResult struct {
	SourcePath string
	// Outputsは書き込みに成功したパスの一覧を保持する。
	Outputs []string
	// Errは変換の一部でも失敗した場合にnil以外になる。
	Err error
}

// Succeededはその変換元ファイルの変換全体がエラーなく完了したかどうかを返す。
func (r ConversionResult) Succeeded() bool { return r.Err == nil }
