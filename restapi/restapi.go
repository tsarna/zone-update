package restapi

import (
  "encoding/json"
  "fmt"
  "github.com/go-chi/chi"
  "github.com/go-chi/chi/middleware"
  "log"
  "net/http"
  "time"
  "zoneupdated/config"
  "zoneupdated/httperror"
  "zoneupdated/updater"
)

type RestApi struct {
  conf    config.Config
  updater updater.Updater
  creds    map[string]string
}

func ServeHttp(conf config.Config, updater updater.Updater) {
  api := RestApi{ conf: conf, updater: updater, creds: make(map[string]string) }

  if conf.User != "" && conf.Password != "" {
     api.creds[conf.User] = conf.Password
  }

  r := chi.NewRouter()

  r.Use(middleware.RequestID)
  r.Use(middleware.RealIP)
  r.Use(middleware.Logger)
  r.Use(middleware.Recoverer)

  if len(api.creds) > 0 {
    r.Use(middleware.BasicAuth(conf.HttpAuthRealm, api.creds))
  }

  // Set a timeout value on the request context (ctx), that will signal
  // through ctx.Done() that the request has timed out and further
  // processing should be stopped.
  r.Use(middleware.Timeout(time.Second * time.Duration(conf.HttpTimeoutSecs)))

  r.Route("/zone-update", func(r chi.Router) {
    r.Post("/present", api.presentEntry)
    r.Post("/cleanup", api.disableEntry)
  })

  log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", conf.HttpAddr, conf.HttpPort), r))
}

func (api *RestApi) presentEntry(w http.ResponseWriter, r *http.Request) {
  api.updateEntry(w, r, false)
}

func (api *RestApi) disableEntry(w http.ResponseWriter, r *http.Request) {
  api.updateEntry(w, r, true)
}

func (api *RestApi) updateEntry(w http.ResponseWriter, r *http.Request, disable bool) {
  updateRequest := updater.UpdateRequest{RRType: "TXT"}

  decoder := json.NewDecoder(r.Body)
  if err := decoder.Decode(&updateRequest); err != nil {
    http.Error(w, fmt.Sprint("JSON Parse error: ", err), http.StatusBadRequest)
    return
  }

  updateRequest.Disable = disable

  if updateRequest.FQDN == "" {
    http.Error(w, "fqdn not provided", http.StatusBadRequest)
    return
  }

  if updateRequest.Value == "" {
    http.Error(w, "value not provided", http.StatusBadRequest)
    return
  }

  err := api.updater.Update(r.Context(), updateRequest)
  if err != nil {
    switch s := err.(type) {
    case httperror.HttpError:
      http.Error(w, s.Error(), s.HttpStatus())
    default:
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  } else {
    _, _ = w.Write([]byte("OK"))
  }
}