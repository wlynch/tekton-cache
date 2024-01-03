package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gorilla/mux"
)

var (
	dir       = os.TempDir()
	nameRegex = regexp.MustCompile(`^[a-zA-Z0-9]{1,128}$`)
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{id}", readItem).Methods("GET")
	r.HandleFunc("/{id}", put).Methods("PUT")
	http.ListenAndServe(":8080", r)
}

func put(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := write(id, r.Body); err != nil {
		http.Error(w, "error writing file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func readItem(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := read(id, w); err != nil {
		http.Error(w, "error reading file", http.StatusNotFound)
		return
	}
}

func write(name string, r io.Reader) error {
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("invalid name")
	}

	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}

func read(name string, w io.Writer) error {
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("invalid name")
	}

	f, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}
