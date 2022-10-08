package resize

type Resizer interface {
	ResizeImage(data []byte, width uint, height uint) ([]byte, error)
}
