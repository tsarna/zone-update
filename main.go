package main

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gofrs/flock"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"zoneupdated/atomicfile"
)

// {
//  "fqdn": "_acme-challenge.domain.",
//  "value": "LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM"
//}

type UpdateRequest struct {
	FQDN   string `json:"fqdn"`
	RRType string `json:"rrtype"`
	Value  string `json:"value"`
}

var zoneFileName string

func main() {
	var port int
	var httpTimeout int

	flag.IntVar(&port, "port", 8080, "HTTP port to listen on")
	flag.IntVar(&httpTimeout, "http-timeout", 60, "HTTP Request timeout")

	flag.Usage = usage
	flag.Parse()

	if (len(flag.Args()) != 1) {
		flag.Usage()
		return
	}

	zoneFileName = flag.Arg(0)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(time.Second * time.Duration(httpTimeout)))

	r.Route("/zone-update", func(r chi.Router) {
		r.Post("/present", presentEntry)
		r.Post("/cleanup", disableEntry)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}

func presentEntry(w http.ResponseWriter, r *http.Request) {
	updateEntry(w, r, true)
}

func disableEntry(w http.ResponseWriter, r *http.Request) {
	updateEntry(w, r, false)
}

func updateEntry(w http.ResponseWriter, r *http.Request, enabled bool) {
	update := UpdateRequest{RRType: "TXT"}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&update); err != nil {
		http.Error(w, fmt.Sprint("JSON Parse error: ", err), http.StatusBadRequest)
		return
	}

	fileLock := flock.New(fmt.Sprintf("%s.lock", zoneFileName))
  success, err := fileLock.TryLockContext(r.Context(), time.Second)
  if success {
		updateZoneFile(w, zoneFileName, update, enabled);

		fileLock.Unlock()
	} else {
		if (err == nil) {
			err = fmt.Errorf("unknown error")
		}

		http.Error(w, fmt.Sprintf("Failed to lock zone file: ", err), http.StatusConflict)
	}
}

func updateZoneFile(w http.ResponseWriter, zoneFileName string, request UpdateRequest, enabled bool) {
	zoneFile, err := os.Open(zoneFileName)
	if (err != nil) {
		http.Error(w, fmt.Sprint("Unable to open zone file: ", err), http.StatusInternalServerError)
		return
	}

	newZoneFile, err := atomicfile.Open(zoneFileName)
	if (err != nil) {
		http.Error(w, fmt.Sprint("Unable to open temp file: ", err), http.StatusInternalServerError)
		zoneFile.Close()
		return
	}

	copyAndUpdate(w, zoneFile, newZoneFile, request, enabled)

	zoneFile.Close()
}

func copyAndUpdate(w http.ResponseWriter, currentFile io.Reader, newFile *atomicfile.AtomicFile,
	request UpdateRequest, enabled bool) {


	newFile.Abort()
	w.Write([]byte("OK"))
}

func cNameHash(fqdn string) string {
	sum := sha1.Sum([]byte(fqdn))
	return base32.StdEncoding.EncodeToString(sum[:])
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s zone-file-name\n\n", os.Args[0])
	flag.PrintDefaults()
}