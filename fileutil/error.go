package fileutil

type InvalidTypeError struct {
	Filename string
}

func newInvalidType(filename string) *InvalidTypeError {
	return &InvalidTypeError{filename}
}

func (e *InvalidTypeError) Error() string {
	return "invalid file type: " + e.Filename
}
