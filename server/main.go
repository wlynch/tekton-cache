package main

import (
	"encoding/json"
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
	// Init data dirs
	for _, d := range []string{"blobs", "metadata"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0700); err != nil {
			panic(err)
		}
	}

	r := mux.NewRouter()

	r.HandleFunc("/{id}", getMetadataHandler).Methods("GET")
	r.HandleFunc("/{id}/blob", getBlobHandler).Methods("GET")

	r.HandleFunc("/{id}", putBlobHandler).Methods("PUT")
	http.ListenAndServe(":8080", r)
}

func putBlobHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := writeBlob(id, r.Body); err != nil {
		http.Error(w, fmt.Sprintf("error writing file: %v", err), http.StatusInternalServerError)
		return
	}
	md := map[string]any{
		"type": r.Header.Get("Content-Type"),
		"size": r.Header.Get("Content-Length"),
	}
	for k, v := range r.URL.Query() {
		md[k] = v
	}
	writeMetadata(id, md)
	w.WriteHeader(http.StatusCreated)
}

func getBlobHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := read(id, w); err != nil {
		http.Error(w, fmt.Sprintf("error reading file: %v", err), http.StatusNotFound)
		return
	}
}

func getMetadataHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := readMetadata(id, w); err != nil {
		http.Error(w, "error reading file", http.StatusNotFound)
		return
	}
}

func writeBlob(name string, r io.Reader) error {
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("invalid name %q", name)
	}

	f, err := os.Create(filepath.Join(dir, "blobs", name))
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

	f, err := os.Open(filepath.Join(dir, "blobs", name))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}

func writeMetadata(name string, data any) error {
	f, err := os.Create(filepath.Join(dir, "metadata", name))
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(data)
}

func readMetadata(name string, w io.Writer) error {
	f, err := os.Open(filepath.Join(dir, "metadata", name))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}
