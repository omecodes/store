package files

import (
	"net/http"
	"os"
)

func Mime(filename string) (contentType string) {
	contentType = "text/plain"

	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	buffer := make([]byte, 512)

	_, err = f.Read(buffer)
	if err != nil {
		return
	}

	contentType = http.DetectContentType(buffer)
	return
}
