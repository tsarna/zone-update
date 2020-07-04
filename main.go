package main

import (
	"context"
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
var lockTimeout int

func main() {
	var port int
	var httpTimeout int

	flag.IntVar(&port, "port", 8080, "HTTP port to listen on")
	flag.IntVar(&httpTimeout, "http-timeout", 60, "HTTP Request timeout")
	flag.IntVar(&lockTimeout, "lock-timeout", 30, "Seconds to wait to obtain a lock")

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
  err := lockWithTimeout(fileLock, lockTimeout, r.Context())
  if (err != nil) {
		http.Error(w, fmt.Sprintf("Failed to lock zone file after %d seconds: %s", lockTimeout, err),
			http.StatusConflict)
	}

  updateZoneFile(w, zoneFileName, update, enabled);

	fileLock.Unlock()
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

func lockWithTimeout(lock *flock.Flock, wait int, ctx context.Context) error {
	var err error

	for tries := wait; tries > 0; tries-- {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = lock.Lock();
		}

		if err == nil {
			return nil
		} else {
			time.Sleep(time.Second)
		}
	}

	return err
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s zone-file-name\n\n", os.Args[0])
	flag.PrintDefaults()
}