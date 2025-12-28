package encoder

import (
	"fmt"
	"io"
	"net/http"
)

func WriteText(w http.ResponseWriter, status int, data string) error {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	_, err := fmt.Fprint(w, data)
	return err
}

func ReadText(r *http.Request, data string) error {
	body, err := io.ReadAll(r.Body)
	data = string(body)
	return err
}
