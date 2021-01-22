package oms

import (
	"github.com/gorilla/mux"
	"github.com/omecodes/store/objects"
	"net/http"
)

func NewHttpUnit() *HTTPUnit {
	return &HTTPUnit{}
}

type HTTPUnit struct{}

func (s *HTTPUnit) MuxRouter() *mux.Router {
	r := mux.NewRouter()
	r.PathPrefix("/objects").Handler(http.StripPrefix("/objects", objects.NewHTTPRouter()))
	return r
}
