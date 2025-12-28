package telemetry

import (
	"encoding/binary"
	"log"
	"net/http"
	"strconv"

	"github.com/sturdivant20/sturdr-api/include/encoder"
)

type Handler struct {
	service Service
}

func NewHttpHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var data Telemetry
	if err := readRequest(r, &data); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// create telemetry using service
	if err := h.service.createTelemetry(r.Context(), data); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// send success status
	encoder.WriteText(w, http.StatusOK, "Success")
}

// Handle http read specific telemetry json request
func (h *Handler) Read(w http.ResponseWriter, r *http.Request) {
	// request telemetry by gps week and tow
	week, tow, do_query := parseQuery(r)

	// read queried telemetry from table
	data, err := h.service.readTelemetry(r.Context(), week, tow, do_query)
	if err != nil {
		handleError(w, err, http.StatusNotFound)
		return
	}

	if err := writeResponse(w, r, data); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}
}

// Handle http update specific telemetry request
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	// request telemetry by id
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// decode incoming json data
	var data Telemetry
	if err := readRequest(r, &data); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// update held telemetry
	if err := h.service.updateTelemetry(r.Context(), data, id); err != nil {
		handleError(w, err, http.StatusExpectationFailed)
		return
	}

	// send success status
	encoder.WriteText(w, http.StatusOK, "Success")
}

// Handle http delete specific telemetry request
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	// request telemetry by id
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// delete held telemetry
	if err := h.service.deleteTelemetry(r.Context(), id); err != nil {
		handleError(w, err, http.StatusExpectationFailed)
		return
	}

	// send success status
	encoder.WriteText(w, http.StatusOK, "Success")
}

// Reusable error handler
func handleError(w http.ResponseWriter, e error, c int) {
	log.Printf("Telemetry error %d! %s", c, e.Error())
	http.Error(w, e.Error(), c)
}

// Reusable read function
func readRequest(r *http.Request, data *Telemetry) error {
	switch r.URL.Query().Get("format") {
	case "binary":
		if err := encoder.ReadBinary(r, data); err != nil {
			return err
		}
	case "json":
		fallthrough
	default:
		if err := encoder.ReadJson(r, data); err != nil {
			return err
		}
	}

	return nil
}

// Reusable write function
func writeResponse(w http.ResponseWriter, r *http.Request, data []Telemetry) error {
	switch r.URL.Query().Get("format") {
	case "binary":
		// write binary data (because a slice is used, the "WriteBinary" function cannot be used)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		for i := 0; i < len(data); i++ {
			if err := binary.Write(w, binary.BigEndian, data[i].Navigation); err != nil {
				return err
			}
			for _, sv := range data[i].Satellites {
				if err := binary.Write(w, binary.BigEndian, sv); err != nil {
					return err
				}
			}
		}
	case "json":
		fallthrough
	default:
		if err := encoder.WriteJson(w, http.StatusOK, data); err != nil {
			return err
		}
	}

	return nil
}

func parseQuery(r *http.Request) (uint16, float32, bool) {
	query := r.URL.Query()
	s_week := query.Get("week")
	s_tow := query.Get("tow")
	if s_week == "" || s_tow == "" {
		return 0, 0.0, false
	}

	week, err1 := strconv.ParseUint(s_week, 10, 16)
	tow, err2 := strconv.ParseFloat(s_tow, 32)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}

	return uint16(week), float32(tow), true
}
