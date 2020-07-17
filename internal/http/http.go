package http

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"

	"github.com/r-erema/wapi/internal/config"
	jsonInfra "github.com/r-erema/wapi/internal/infrastructure/json"
	"github.com/r-erema/wapi/internal/infrastructure/os"
	"github.com/r-erema/wapi/internal/repository"
	"github.com/r-erema/wapi/internal/service"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// AppError is custom http application error
type AppError struct {
	Error       error
	ResponseMsg string
	Code        int
}

// AppHTTPHandler is custom http handler returning custom application error type
type AppHTTPHandler interface {
	Handle(http.ResponseWriter, *http.Request) *AppError
}

// AppHandlerRunner starts application http handler implemented AppHTTPHandler interface
type AppHandlerRunner struct {
	H AppHTTPHandler
}

// ServeHTTP handles http request of particular AppHTTPHandler interface implementation.
func (fn AppHandlerRunner) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn.H.Handle(w, r); err != nil {
		http.Error(w, err.ResponseMsg, err.Code)
		log.Printf("http request serving error: %+v\n", err.Error)
	}
}

// Router creates http handlers and bind them with paths.
func Router(
	conf *config.Config,
	sessRepo repository.Session,
	connSupervisor service.Connections,
	authorizer service.Authorizer,
	qrFileResolver service.QRFileResolver,
	listener service.Listener,
	fs os.FileSystem,
) (*mux.Router, error) {
	if conf.Env == config.DevMode {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, // nolint
		}
	}

	registerHandler := NewRegisterSessionHandler(authorizer, listener, sessRepo)
	log.Print("trying to auto connect saved sessions if exist...")
	if err := registerHandler.TryToAutoConnectAllSessions(); err != nil {
		return nil, err
	}
	marshal := jsonInfra.MarshallCallback(json.Marshal)
	sendMessageHandler := NewTextHandler(authorizer, connSupervisor, &marshal)
	sendImageHandler := NewImageHandler(authorizer, connSupervisor, &http.Client{}, &marshal)
	getQRImageHandler := NewQR(fs, qrFileResolver)
	getSessionInfoHandler := NewSessInfoHandler(sessRepo)
	getActiveConnectionInfoHandler := NewInfo(connSupervisor)

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"Content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)
	router := mux.NewRouter().StrictSlash(true)
	router.Use(cors)

	router.Handle("/register-session/", AppHandlerRunner{H: registerHandler}).Methods(http.MethodPost)
	router.Handle("/send-message/", AppHandlerRunner{H: sendMessageHandler}).Methods(http.MethodPost)
	router.Handle("/send-image/", AppHandlerRunner{H: sendImageHandler}).Methods(http.MethodPost)
	router.Handle("/get-qr-code/{sessionID}/", AppHandlerRunner{H: getQRImageHandler}).Methods(http.MethodGet)
	router.Handle("/get-session-info/{sessionID}/", AppHandlerRunner{H: getSessionInfoHandler}).Methods(http.MethodGet)
	router.Handle("/get-active-connection-info/{sessionID}/", AppHandlerRunner{H: getActiveConnectionInfoHandler}).Methods(http.MethodGet)

	return router, nil
}
