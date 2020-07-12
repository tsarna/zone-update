package restapi

import (
  "bufio"
  "crypto/tls"
  "encoding/json"
  "fmt"
  "github.com/go-chi/chi"
  "github.com/go-chi/chi/middleware"
  "io"
  "log"
  "net/http"
  "os"
  "strings"
  "sync/atomic"
  "time"
  "zoneupdated/config"
  "zoneupdated/httperror"
  "zoneupdated/updater"
)

type RestApi struct {
  conf    config.Config
  updater updater.Updater
  creds   map[string]string
  cert    atomic.Value
}

func ServeHttp(conf config.Config, updater updater.Updater) error {
  api := RestApi{conf: conf, updater: updater, creds: make(map[string]string)}

  if conf.HttpAuthFile != "" {
    err := api.parseAuthUsers(conf.HttpAuthFile)
    if err != nil {
      return fmt.Errorf("while parsing auth file: %s", err)
    }
  }

  if conf.User != "" && conf.Password != "" {
    api.creds[conf.User] = conf.Password
  }

  r := chi.NewRouter()

  r.Use(middleware.RequestID)

  if conf.TrustProxy {
    r.Use(middleware.RealIP)
  }

  r.Use(middleware.Logger)
  r.Use(middleware.Recoverer)

  // Second test ensures auth is enabled even if auth file is empty, to fail secure
  if len(api.creds) > 0 || conf.HttpAuthFile != "" {
    r.Use(middleware.BasicAuth(conf.HttpAuthRealm, api.creds))
  }

  // Set a timeout value on the request context (ctx), that will signal
  // through ctx.Done() that the request has timed out and further
  // processing should be stopped.
  r.Use(middleware.Timeout(time.Second * time.Duration(conf.HttpTimeoutSecs)))

  r.Route(conf.UrlPrefix, func(r chi.Router) {
    r.Post("/present", api.presentEntry)
    r.Post("/cleanup", api.disableEntry)
  })

  if conf.RobotsTxt {
    r.Get("/robots.txt", robotsTxt)
  }

  if conf.TlsCertFilename != "" && conf.TlsKeyFilename != "" {
    err := api.loadCert()
    if err != nil {
      return nil
    }

    tlsConfig := &tls.Config{
      GetCertificate: api.getCertificate,
    }
    server := &http.Server{
      Addr:      conf.ListenAddr,
      Handler:   r,
      TLSConfig: tlsConfig,
    }
    log.Fatal(server.ListenAndServeTLS("", ""))
  } else {
    log.Fatal(http.ListenAndServe(conf.ListenAddr, r))
  }

  return nil
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

func (api *RestApi) parseAuthUsers(httpAuthFile string) error {
  file, err := os.Open(httpAuthFile)
  if err != nil {
    return err
  }
  defer file.Close()

  scanner := bufio.NewScanner(file)
  scanner.Split(bufio.ScanLines)
  linenumber := 1

  for scanner.Scan() {
    line := scanner.Text()
    fields := strings.Fields(line)

    if len(fields) == 0 {
      continue
    } else if len(fields) == 2 {
      api.creds[fields[0]] = fields[1]
    } else {
      return fmt.Errorf("Line needs two fields at line %d: '%s'", linenumber, line)
    }

    linenumber++
  }

  return nil
}

func (api *RestApi) loadCert() error {
  cert, err := tls.LoadX509KeyPair(api.conf.TlsCertFilename, api.conf.TlsKeyFilename)
  if err != nil {
    return err
  }

  api.cert.Store(&cert)

  return nil
}

func (api *RestApi) getCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
  cert, ok := api.cert.Load().(*tls.Certificate)

  if cert == nil || !ok {
    return nil, fmt.Errorf("No valid certificate loaded")
  }

  return cert, nil
}

func robotsTxt(w http.ResponseWriter, _ *http.Request) {
  _, _ = io.WriteString(w, "User-agent: *\n")
  _, _ = io.WriteString(w, "Disallow: /")
}
