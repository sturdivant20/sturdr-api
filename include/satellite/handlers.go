package satellite

import (
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

// Handle http create satellite json request
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var sv Satellite
	if err := readRequest(r, &sv); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// create satellite using service
	if err := h.service.createSatellite(r.Context(), sv); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// send success status
	encoder.WriteText(w, http.StatusOK, "Success")
}

// Handle http read specific satellite json request
func (h *Handler) Read(w http.ResponseWriter, r *http.Request) {
	// request navigation by gps week, tow, and prn
	week, tow, prn, do_query := parseQuery(r)

	// read specific satellite from table
	sv, err := h.service.readSatellite(r.Context(), week, tow, prn, do_query)
	if err != nil {
		handleError(w, err, http.StatusNotFound)
		return
	}

	if err := writeResponse(w, r, sv); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}
}

// Handle http update specific satellite request
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	// request satellite by id
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// decode incoming json data
	var sv Satellite
	if err := readRequest(r, &sv); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// update held satellite
	if err := h.service.updateSatellite(r.Context(), sv, id); err != nil {
		handleError(w, err, http.StatusExpectationFailed)
		return
	}

	// send success status
	encoder.WriteText(w, http.StatusOK, "Success")
}

// Handle http delete specific satellite reques
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	// request satellite by id
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// delete held satellite
	if err := h.service.deleteSatellite(r.Context(), id); err != nil {
		handleError(w, err, http.StatusExpectationFailed)
		return
	}

	// send success status
	encoder.WriteText(w, http.StatusOK, "Success")
}

// Reusable error handler
func handleError(w http.ResponseWriter, e error, c int) {
	log.Printf("Satellite error %d! %s", c, e.Error())
	http.Error(w, e.Error(), c)
}

// Reusable read function
func readRequest(r *http.Request, sv *Satellite) error {
	switch r.URL.Query().Get("format") {
	case "binary":
		if err := encoder.ReadBinary(r, sv); err != nil {
			return err
		}
	case "json":
		fallthrough
	default:
		if err := encoder.ReadJson(r, sv); err != nil {
			return err
		}
	}

	return nil
}

// Reusable write function
func writeResponse(w http.ResponseWriter, r *http.Request, sv []Satellite) error {
	switch r.URL.Query().Get("format") {
	case "binary":
		if err := encoder.WriteBinary(w, http.StatusOK, sv); err != nil {
			return err
		}
	case "json":
		fallthrough
	default:
		if err := encoder.WriteJson(w, http.StatusOK, sv); err != nil {
			return err
		}
	}

	return nil
}

func parseQuery(r *http.Request) (uint16, float32, uint8, bool) {
	query := r.URL.Query()
	s_week := query.Get("week")
	s_tow := query.Get("tow")
	s_prn := query.Get("prn")
	var prn uint64
	var err3 error
	if s_prn == "" {
		prn = 255
	} else {
		prn, err3 = strconv.ParseUint(s_prn, 10, 8)
		if err3 != nil {
			prn = 255
		}
	}
	if s_week == "" || s_tow == "" {
		return 0, 0.0, uint8(prn), false
	}

	week, err1 := strconv.ParseUint(s_week, 10, 16)
	tow, err2 := strconv.ParseFloat(s_tow, 32)
	if err1 != nil || err2 != nil {
		return 0, 0, uint8(prn), false
	}

	return uint16(week), float32(tow), uint8(prn), true
}
