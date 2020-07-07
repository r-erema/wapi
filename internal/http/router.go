package http

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"

	"github.com/r-erema/wapi/internal/config"
	jsonInfra "github.com/r-erema/wapi/internal/infrastructure/json"
	"github.com/r-erema/wapi/internal/repository"
	"github.com/r-erema/wapi/internal/service/auth"
	msg "github.com/r-erema/wapi/internal/service/message"
	"github.com/r-erema/wapi/internal/service/qr/file"
	"github.com/r-erema/wapi/internal/service/supervisor"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Router creates http handlers and bind them with paths.
func Router(
	conf *config.Config,
	sessRepo repository.SessionRepository,
	connSupervisor supervisor.Connections,
	authorizer auth.Authorizer,
	qrFileResolver file.QRFileResolver,
	listener msg.Listener,
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
	sendMessageHandler := NewTextHandler(authorizer, connSupervisor)
	marshal := jsonInfra.MarshallCallback(json.Marshal)
	sendImageHandler := NewImageHandler(authorizer, connSupervisor, &http.Client{}, &marshal)
	getQRImageHandler := NewQR(qrFileResolver)
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
