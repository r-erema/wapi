package http

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/r-erema/wapi/internal/config"
	jsonInfra "github.com/r-erema/wapi/internal/infrastructure/json"
	"github.com/r-erema/wapi/internal/infrastructure/os"
	"github.com/r-erema/wapi/internal/repository"
	"github.com/r-erema/wapi/internal/service"
)

// Router creates http handlers and bind them with paths.
func Router(
	conf *config.Config,
	sessRepo repository.SessionRepository,
	connSupervisor service.Connections,
	authorizer service.Authorizer,
	qrFileResolver service.QRFileResolver,
	listener service.Listener,
	fs os.FileSystem,
) *mux.Router {
	if conf.Env == config.DevMode {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, // nolint
		}
	}

	registerHandler := NewRegisterSessionHandler(authorizer, listener, sessRepo)
	log.Print("trying to auto connect saved sessions if exist...")
	if err := registerHandler.TryToAutoConnectAllSessions(); err != nil {
		log.Fatalf("error while trying restore sesssions: %s", err)
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

	router.Handle("/register-session/", registerHandler).Methods("POST")
	router.Handle("/send-message/", sendMessageHandler).Methods("POST")
	router.Handle("/send-image/", sendImageHandler).Methods("POST")
	router.Handle("/get-qr-code/{sessionID}/", getQRImageHandler).Methods("GET")
	router.Handle("/get-session-info/{sessionID}/", getSessionInfoHandler).Methods("GET")
	router.Handle("/get-active-connection-info/{sessionID}/", getActiveConnectionInfoHandler).Methods("GET")

	return router
}
