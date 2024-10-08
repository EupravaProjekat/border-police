package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/EupravaProjekat/border-police/Models"
	"github.com/EupravaProjekat/border-police/Repo"
	"github.com/google/uuid"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
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
		err := errors.New("user doesnt exist")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)

}
func (h *Borderhendler) NewUser(w http.ResponseWriter, r *http.Request) {

	res := ValidateJwt2(r, h.repo)
	res2 := ValidateJwt(r, h.repo)
	rt, err := DecodeBodyUser(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusAccepted)
		return
	}
	if res != rt.Email {
		err := errors.New("user doesnt exist")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	newUUID := uuid.New().String()
	rt.Uuid = newUUID
	rt.Role = "Operator"
	rt.Email = res2.Email
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
func (h *Borderhendler) NewCausing(w http.ResponseWriter, r *http.Request) {

	_ = ValidateJwt(r, h.repo)

	if r.Header.Get("intern") != "prosecution-service-secret-code" {
		err := errors.New("not allowed")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	rt, err := VehicleCausing(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusAccepted)
		return
	}
	newUUID := uuid.New().String()
	rt.Status = "new"
	rt.Uuid = newUUID
	err = h.repo.NewCausing(rt)
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
func (h *Borderhendler) GetallCausings(w http.ResponseWriter, r *http.Request) {

	res := ValidateJwt(r, h.repo)
	if res == nil {
		err := errors.New("user doesnt exist")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if res.Role != "Operator" {
		err := errors.New("role error")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	response, err := h.repo.GetAllCausings()
	if err != nil {
		log.Printf("Operation Failed: %v\n", err)
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("Casings not found"))
		if err != nil {
			return
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	RenderJSON(w, response)
}
func (h *Borderhendler) GetallRequests(w http.ResponseWriter, r *http.Request) {

	res := ValidateJwt(r, h.repo)
	if res == nil {
		err := errors.New("user doesnt exist")
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
		w.WriteHeader(http.StatusNotFound)
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

	res := ValidateJwt(r, h.repo)
	if res == nil {
		err := errors.New("user doesnt exist")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)
	RenderJSON(w, res)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newUUID := uuid.New().String()
	rt.Uuid = newUUID
	rt.Status = "received"

	userData := []byte(`{"plates":"` + rt.CarPlateNumber + `"}`)

	apiUrl := "http://police-service:9099/checkplateswanted"
	request, err2 := http.NewRequest("POST", apiUrl, bytes.NewBuffer(userData))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("jwt", r.Header.Get("jwt"))
	request.Header.Set("intern", "border-service-secret-code")

	// send the request
	client := &http.Client{}
	response, err2 := client.Do(request)
	if err2 != nil {
		fmt.Println(err2)
	}

	responseBody, err2 := io.ReadAll(response.Body)
	if err2 != nil {
		fmt.Println(err2)
	}

	var resp Models.Response
	err = json.Unmarshal(responseBody, &resp)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Extract the boolean value.
	b := resp.VehicleWanted
	fmt.Println(b)
	rt.VehicleWanted = strconv.FormatBool(b)
	defer response.Body.Close()

	res := ValidateJwt(r, h.repo)
	if res == nil {
		err := errors.New("user doesnt exist")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	re := res
	re.Requests = append(re.Requests, *rt)
	err = h.repo.Update(re)
	if err != nil {
		log.Printf("Operation failed: %v\n", err)
		w.WriteHeader(http.StatusNotFound)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res := ValidateJwt(r, h.repo)
	if res == nil {
		err := errors.New("user doesnt exist")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	respon, err := h.repo.GetRequest(rt.Uuid)
	if err != nil {
		log.Printf("Operation failed: %v\n", err)
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("couldn't find request"))
		if err != nil {
			return
		}
		return
	}
	RenderJSON(w, respon)
	w.WriteHeader(http.StatusOK)
}
func (h *Borderhendler) UpdateRequest(w http.ResponseWriter, r *http.Request) {

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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res := ValidateJwt(r, h.repo)
	if res == nil {
		err := errors.New("user doesnt exist")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	respon, err := h.repo.UpdateRequest(rt.Uuid)
	if err != nil {
		log.Printf("Operation failed: %v\n", err)
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("couldn't find request"))
		if err != nil {
			return
		}
		return
	}
	RenderJSON(w, respon)
	w.WriteHeader(http.StatusOK)
}
