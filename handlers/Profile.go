package handlers

import (
	"errors"
	"github.com/EupravaProjekat/border-police/Repo"
	protos "github.com/MihajloJankovic/profile-service/protos/main"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"mime"
	"net/http"
)

type Borderhendler struct {
	l    *log.Logger
	repo *Repo.Repo
}

func NewBorderhendler(l *log.Logger, r *Repo.Repo) *Borderhendler {
	return &Borderhendler{l, r}

}

func (h *Borderhendler) CheckIfUserExists(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		err := errors.New("expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	res := ValidateJwt(r, h.repo)
	if res == nil {
		err := errors.New("jwt error")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	re := res
	response, err := h.repo.GetByEmail(re.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if re.Email != response.Email {
		err := errors.New("authorization error")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

}
func (h *Borderhendler) NewUser(w http.ResponseWriter, r *http.Request) {

	rt, err := DecodeBodyUser(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusAccepted)
		return
	}
	newUUID := uuid.New().String()
	rt.Uuid = newUUID
	rt.Role = "Guest"
	err = h.repo.NewUser(rt)
	if err != nil {
		log.Printf("Operation Failed: %v\n", err)
		w.WriteHeader(http.StatusNotAcceptable)
		_, err := w.Write([]byte("Profile not found"))
		if err != nil {
			return
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (h *Borderhendler) GetallRequests(w http.ResponseWriter, r *http.Request) {

	res := ValidateJwt(r, h.repo)
	if res == nil {
		err := errors.New("jwt error")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if res.Role != "Operator" {
		err := errors.New("role error")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	response, err := h.repo.GetAllRequest()
	if err != nil {
		log.Printf("Operation Failed: %v\n", err)
		w.WriteHeader(http.StatusNotAcceptable)
		_, err := w.Write([]byte("Requests not found"))
		if err != nil {
			return
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	RenderJSON(w, response)
}
func (h *Borderhendler) GetProfile(w http.ResponseWriter, r *http.Request) {

	emaila := mux.Vars(r)["email"]
	ee := new(protos.ProfileRequest)
	ee.Email = emaila
	res := ValidateJwt(r, h.repo)
	if res == nil {
		err := errors.New("jwt error")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	response, err := h.repo.GetByEmail(ee.Email)
	if err != nil || response == nil {
		log.Printf("Operation Failed: %v\n", err)
		w.WriteHeader(http.StatusNotAcceptable)
		_, err := w.Write([]byte("Profile not found"))
		if err != nil {
			return
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	RenderJSON(w, response)
}

func (h *Borderhendler) NewRequest(w http.ResponseWriter, r *http.Request) {

	contentType := r.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		err := errors.New("expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	rt, err := DecodeBody(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusAccepted)
		return
	}
	newUUID := uuid.New().String()
	rt.Uuid = newUUID
	rt.Status = "received"
	res := ValidateJwt(r, h.repo)
	if res == nil {
		err := errors.New("jwt error")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	re := res
	re.Requests = append(re.Requests, *rt)
	err = h.repo.Update(re)
	if err != nil {
		log.Printf("Operation failed: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("couldn't add request"))
		if err != nil {
			return
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("successfully added request"))
	if err != nil {
		return
	}
}
func (h *Borderhendler) GetRequest(w http.ResponseWriter, r *http.Request) {

	contentType := r.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		err := errors.New("expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	rt, err := DecodeBody2(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusAccepted)
		return
	}
	res := ValidateJwt(r, h.repo)
	if res == nil {
		err := errors.New("jwt error")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	respon, err := h.repo.GetRequest(rt.Uuid)
	if err != nil {
		log.Printf("Operation failed: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("couldn't add request"))
		if err != nil {
			return
		}
		return
	}
	RenderJSON(w, respon)
	w.WriteHeader(http.StatusOK)
}
