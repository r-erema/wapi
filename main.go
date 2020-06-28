package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/r-erema/wapi/internal/config"
	"github.com/r-erema/wapi/internal/http/handler/connection"
	"github.com/r-erema/wapi/internal/http/handler/message"
	"github.com/r-erema/wapi/internal/http/handler/qr"
	"github.com/r-erema/wapi/internal/http/handler/session"
	messageRepo "github.com/r-erema/wapi/internal/repository/message"
	sessionRepo "github.com/r-erema/wapi/internal/repository/session"
	"github.com/r-erema/wapi/internal/service/auth"
	"github.com/r-erema/wapi/internal/service/listener"
	"github.com/r-erema/wapi/internal/service/supervisor"

	_ "github.com/Rhymen/go-whatsapp"
	"github.com/getsentry/sentry-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	conf, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("connecting to redis...")
	redisClient, err := messageRepo.NewRedis(conf.RedisHost)
	if err != nil {
		log.Fatalf("error of init redis repo: %v\n", err)
	}
	log.Print("ok")

	if conf.SentryDSN != "" {
		log.Print("init Sentry...")
		err = initSentry(conf.SentryDSN)
		if err != nil {
			log.Fatalf("error of init sentry: %v\n", err)
		}
		log.Print("ok")
	}

	sessionWorks, err := sessionRepo.NewFileSystem(conf.FileSystemRootPath + "/sessions")
	if err != nil {
		log.Fatalf("can't create service `session`: %v\n", err)
	}

	connSupervisor := supervisor.New(time.Duration(conf.ConnectionsCheckoutDuration))

	a, err := auth.New(
		conf.FileSystemRootPath+"/qr-codes",
		time.Duration(conf.ConnectionTimeout)*time.Second,
		sessionWorks,
		connSupervisor,
	)
	if err != nil {
		log.Fatalf("can't create service `a`: %v\n", err)
	}

	l := listener.NewWebHook(sessionWorks, connSupervisor, a, conf.WebHookURL, redisClient)

	registerHandler := session.NewRegisterSessionHandler(a, l, sessionWorks)
	log.Print("trying to auto connect saved sessions if exist...")
	if err = registerHandler.TryToAutoConnectAllSessions(); err != nil {
		log.Fatalf("error while trying restore sesssions: %s", err)
	}

	sendMessageHandler := message.NewTextHandler(a, connSupervisor)
	sendImageHandler := message.NewImageHandler(a, connSupervisor)
	getQRImageHandler := qr.New(a)
	getSessionInfoHandler := session.NewSessInfoHandler(sessionWorks)
	getActiveConnectionInfoHandler := connection.New(connSupervisor)

	err = runServer(
		conf,
		*registerHandler,
		*sendMessageHandler,
		*sendImageHandler,
		*getQRImageHandler,
		*getSessionInfoHandler,
		*getActiveConnectionInfoHandler,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func runServer(
	conf *config.Config,
	registerHandler session.RegisterSessionHandler,
	sendMessageHandler message.SendTextMessageHandler,
	sendImageHandler message.SendImageHandler,
	getQRImageHandler qr.GetQRImageHandler,
	getSessionInfoHandler session.SessInfoHandler,
	getActiveConnectionInfoHandler connection.ActiveConnectionInfoHandler,
) error {
	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"Content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)
	router := mux.NewRouter().StrictSlash(true)
	router.Use(cors)

	router.Handle("/register-session/", &registerHandler).Methods("POST")
	router.Handle("/send-message/", &sendMessageHandler).Methods("POST")
	router.Handle("/send-image/", &sendImageHandler).Methods("POST")
	router.Handle("/get-qr-code/{sessionId}/", &getQRImageHandler).Methods("GET")
	router.Handle("/get-session-info/{sessionId}/", &getSessionInfoHandler).Methods("GET")
	router.Handle("/get-active-connection-info/{sessionId}/", &getActiveConnectionInfoHandler).Methods("GET")

	var err error
	certFileExists, certKeyExists := true, true
	if _, err = os.Stat(conf.CertFilePath); os.IsNotExist(err) {
		certFileExists = false
	}
	if _, err = os.Stat(conf.CertKeyPath); os.IsNotExist(err) {
		certKeyExists = false
	}

	if conf.Env == config.DevMode {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, // nolint
		}
	}

	if !certFileExists || !certKeyExists {
		log.Printf("wapi will handle request by unsecured connection. Wapi's listening at %s ...\n", conf.ListenHTTPHost)
		err = http.ListenAndServe(conf.ListenHTTPHost, router)
	} else {
		log.Printf("wapi's listening at %s ...\n", conf.ListenHTTPHost)
		err = http.ListenAndServeTLS(conf.ListenHTTPHost, conf.CertFilePath, conf.CertKeyPath, router)
	}
	if err != nil {
		return err
	}
	return nil
}

func initSentry(dsn string) error {
	if dsn == "" {
		return fmt.Errorf("senrty dsn couldn't be empty")
	}
	if err := sentry.Init(sentry.ClientOptions{Dsn: dsn}); err != nil {
		return err
	}
	return nil
}
