package navigation

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

// Handle http create navigation json request
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var n Navigation
	if err := readRequest(r, &n); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// create navigation using service
	if err := h.service.createNavigation(r.Context(), n); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// send success status
	encoder.WriteText(w, http.StatusOK, "Success")
}

// Handle http read specific navigation json request
func (h *Handler) Read(w http.ResponseWriter, r *http.Request) {
	// request navigation by gps week and tow
	week, tow, do_query := parseQuery(r)

	// read queried navigation from table
	n, err := h.service.readNavigation(r.Context(), week, tow, do_query)
	if err != nil {
		handleError(w, err, http.StatusNotFound)
		return
	}

	if err := writeResponse(w, r, n); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}
}

// Handle http update specific navigation request
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	// request navigation by id
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// decode incoming json data
	var n Navigation
	if err := readRequest(r, &n); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// update held navigation
	if err := h.service.updateNavigation(r.Context(), n, id); err != nil {
		handleError(w, err, http.StatusExpectationFailed)
		return
	}

	// send success status
	encoder.WriteText(w, http.StatusOK, "Success")
}

// Handle http delete specific navigation request
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	// request navigation by id
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	// delete held navigation
	if err := h.service.deleteNavigation(r.Context(), id); err != nil {
		handleError(w, err, http.StatusExpectationFailed)
		return
	}

	// send success status
	encoder.WriteText(w, http.StatusOK, "Success")
}

// Reusable error handler
func handleError(w http.ResponseWriter, e error, c int) {
	log.Printf("Navigation error %d! %s", c, e.Error())
	http.Error(w, e.Error(), c)
}

// Reusable read function
func readRequest(r *http.Request, n *Navigation) error {
	switch r.URL.Query().Get("format") {
	case "binary":
		if err := encoder.ReadBinary(r, n); err != nil {
			return err
		}
	case "json":
		fallthrough
	default:
		if err := encoder.ReadJson(r, n); err != nil {
			return err
		}
	}

	return nil
}

// Reusable write function
func writeResponse(w http.ResponseWriter, r *http.Request, n []Navigation) error {
	switch r.URL.Query().Get("format") {
	case "binary":
		if err := encoder.WriteBinary(w, http.StatusOK, n); err != nil {
			return err
		}
	case "json":
		fallthrough
	default:
		if err := encoder.WriteJson(w, http.StatusOK, n); err != nil {
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
