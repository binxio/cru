package main

import (
	"encoding/json"
	"github.com/binxio/cru/ref"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"log"
	"net/http"
	"os"
)

type ContainerReferenceUpdateRequest struct {
	CommitMessage   string   `json:"commit-message"`
	ImageReferences []string `json:"image-references"`
}

type ContainerReferenceUpdateResponse struct {
	Files []string `json:"files,omitempty"`
	Hash  string   `json:"commit-sha,omitempty"`
}

func (c Cru) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var hash plumbing.Hash
	var request ContainerReferenceUpdateRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c.CommitMsg = request.CommitMessage
	if c.CommitMsg == "" {
		http.Error(w, "commit message is empty", http.StatusBadRequest)
		return
	}

	c.imageRefs = make(ref.ContainerImageReferences, 0)
	for _, r := range request.ImageReferences {
		r, err := ref.NewContainerImageReference(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		c.imageRefs = append(c.imageRefs, *r)
	}
	if len(c.imageRefs) == 0 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = c.ConnectToRepository(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = c.Walk(Update); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(c.updatedFiles) > 0 {
		log.Printf("INFO: updated a total of %d files", len(c.updatedFiles))
		if hash, err = c.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err = c.Push(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		log.Println("INFO: no files were updated by cru")
		c.updatedFiles = make([]string, 0)
	}
	if body, err := json.Marshal(ContainerReferenceUpdateResponse{
		c.updatedFiles, hash.String()}); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
		return
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *Cru) ListenAndServe() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if c.Port == "" {
		c.Port = os.Getenv("PORT")
	}
	if c.Port == "" {
		c.Port = "8080"
	}

	log.Printf("Listening on port %s", c.Port)
	if err := http.ListenAndServe(":"+c.Port, c); err != nil {
		log.Fatal(err)
	}
}
