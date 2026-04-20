package rhandler

import "net/http"

type Health interface {
	LastCheck(w http.ResponseWriter, r *http.Request)
}

type (
	Category interface {
		Create(w http.ResponseWriter, r *http.Request)
		GetByGUID(w http.ResponseWriter, r *http.Request)
		Update(w http.ResponseWriter, r *http.Request)
		Delete(w http.ResponseWriter, r *http.Request)
		List(w http.ResponseWriter, r *http.Request)
	}

	Product interface {
		Create(w http.ResponseWriter, r *http.Request)
		GetByGUID(w http.ResponseWriter, r *http.Request)
		Update(w http.ResponseWriter, r *http.Request)
		Delete(w http.ResponseWriter, r *http.Request)
		List(w http.ResponseWriter, r *http.Request)
	}
)
