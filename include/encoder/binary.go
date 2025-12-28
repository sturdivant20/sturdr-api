package encoder

import (
	"encoding/binary"
	"net/http"
)

// big-endian is standard in networking

func WriteBinary(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(status)
	return binary.Write(w, binary.BigEndian, data)
}

func ReadBinary(r *http.Request, data any) error {
	return binary.Read(r.Body, binary.BigEndian, data)
}
