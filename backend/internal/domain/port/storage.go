package port

import "io"

// FileStorageはファイル探索とI/Oを抽象化し、usecase層がOSのファイル
// システムに直接触れないようにする。
type FileStorage interface {
	// FindFilesはpathが通常のファイルを指す場合はそのファイル自身を、
	// ディレクトリを指す場合は直下(recursiveがtrueなら再帰的に配下)の
	// ファイルすべてを返す。
	FindFiles(path string, recursive bool) ([]string, error)
	Open(path string) (io.ReadCloser, error)
	// Createは書き込み用にpathを開き、必要に応じて親ディレクトリを作成する。
	// overwriteがfalseで既にファイルが存在する場合、fs.ErrExistを
	// ラップしたエラーを返す。
	Create(path string, overwrite bool) (io.WriteCloser, error)
}
