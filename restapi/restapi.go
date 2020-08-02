package restapi

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	mymiddleware "github.com/tsarna/chi/middleware"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"
	"zone-update/config"
	"zone-update/httperror"
	"zone-update/updater"
)

type RestApi struct {
	conf        *config.Config
	updater     updater.Updater
	credentials map[string]string
	cert        atomic.Value
	passwords   *PasswordFile
}

func New(conf *config.Config, updater updater.Updater) RestApi {
	return RestApi{conf: conf, updater: updater, credentials: make(map[string]string)}
}

func (api *RestApi) ServeHttp() error {
	var err error

	if api.conf.HttpAuthFile != "" {
		api.passwords, err = NewPasswordFile(api.conf.HttpAuthFile)
		if err != nil {
			return fmt.Errorf("while parsing auth file: %s", err)
		}
	}

	if api.conf.User != "" && api.conf.Password != "" {
		api.credentials[api.conf.User] = api.conf.Password
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)

	if api.conf.TrustProxy {
		r.Use(middleware.RealIP)
	}

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(time.Second * time.Duration(api.conf.HttpTimeoutSecs)))

	r.Route(api.conf.UrlPrefix, func(r chi.Router) {
		// Second test ensures auth is enabled even if auth file is empty, to fail secure
		if api.conf.HttpAuthFile != "" {
			r.Use(mymiddleware.BasicAuthWithAuthenticator(api.conf.HttpAuthRealm, api.passwords))
		} else if len(api.credentials) > 0 {
			r.Use(mymiddleware.BasicAuth(api.conf.HttpAuthRealm, api.credentials))
		}

		r.Post("/present", api.presentEntry)
		r.Post("/cleanup", api.disableEntry)
	})

	if api.conf.RobotsTxt {
		r.Get("/robots.txt", robotsTxt)
	}

	if api.conf.UseHttps() {
		err := api.loadCert()
		if err != nil {
			return err
		}

		tlsConfig := &tls.Config{
			GetCertificate: api.getCertificate,
		}
		server := &http.Server{
			Addr:      api.conf.ListenAddr,
			Handler:   r,
			TLSConfig: tlsConfig,
		}
		log.Fatal(server.ListenAndServeTLS("", ""))
	} else {
		log.Fatal(http.ListenAndServe(api.conf.ListenAddr, r))
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
		_, _ = w.Write([]byte("OK\n"))
	}
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
		return nil, fmt.Errorf("no valid certificate loaded")
	}

	return cert, nil
}

func (api *RestApi) Reload() {
	if api.conf.UseHttps() {
		go func() {
			err := api.loadCert()
			if err != nil {
				log.Printf("Failed to reload certificates: %s", err)
			} else {
				log.Printf("Reloaded certificates")
			}
		}()
	}

	if api.passwords != nil {
		go func() {
			err := api.passwords.Reload()
			if err != nil {
				log.Printf("Failed to reload password file: %s", err)
			} else {
				log.Printf("Reloaded auth file")
			}
		}()
	}
}

func robotsTxt(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, "User-agent: *\nDisallow: /\n")
}
