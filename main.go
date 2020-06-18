package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/Rhymen/go-whatsapp"
	"github.com/getsentry/sentry-go"
	"github.com/go-redis/redis"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/r-erema/wapi/config"
	"github.com/r-erema/wapi/src/RouteHandler"
	"github.com/r-erema/wapi/src/Service/Auth"
	"github.com/r-erema/wapi/src/Service/ConnectionsSupervisor"
	"github.com/r-erema/wapi/src/Service/MessageListener"
	"github.com/r-erema/wapi/src/Service/SessionWorks"
)

func main() {

	conf, err := config.Init("./.env")
	if err != nil {
		log.Fatal(err)
	}

	log.Print("connecting to redis...")
	redisClient, err := initRedis(conf.RedisHost)
	if err != nil {
		log.Fatalf("error of init redis client: %v\n", err)
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

	sessionWorks, err := SessionWorks.NewFileSystemSession(conf.FileSystemRootPath + "/sessions")
	if err != nil {
		log.Fatalf("can't create service `SessionWorks`: %v\n", err)
	}

	connSupervisor := ConnectionsSupervisor.NewConnectionsSupervisor(time.Duration(conf.ConnectionsCheckoutDuration))

	auth, err := Auth.NewAuth(conf.FileSystemRootPath+"/qr-codes", time.Duration(conf.ConnectionTimeout)*time.Second, sessionWorks, connSupervisor)
	if err != nil {
		log.Fatalf("can't create service `Auth`: %v\n", err)
	}

	listener := MessageListener.NewListener(sessionWorks, connSupervisor, auth, conf.WebHookUrl, redisClient)

	registerHandler := RouteHandler.NewRegisterSessionHandler(auth, listener, sessionWorks)
	log.Print("trying to auto connect saved sessions if exist...")
	if err = registerHandler.TryToAutoConnectAllSessions(); err != nil {
		log.Fatalf("error while trying restore sesssions: %s", err)
	}

	sendMessageHandler := RouteHandler.NewSendMessageHandler(auth, connSupervisor)
	sendImageHandler := RouteHandler.NewSendImageHandler(auth, connSupervisor)
	getQRImageHandler := RouteHandler.NewGetQRImageHandler(auth)
	getSessionInfoHandler := RouteHandler.NewGetSessionInfoHandler(sessionWorks)
	getActiveConnectionInfoHandler := RouteHandler.NewGetActiveConnectionInfoHandler(connSupervisor)

	err = runServer(
		conf,
		registerHandler,
		sendMessageHandler,
		sendImageHandler,
		getQRImageHandler,
		getSessionInfoHandler,
		getActiveConnectionInfoHandler,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func runServer(
	conf *config.Config,
	registerHandler *RouteHandler.RegisterSessionHandler,
	sendMessageHandler *RouteHandler.SendTextMessageHandler,
	sendImageHandler *RouteHandler.SendImageHandler,
	getQRImageHandler *RouteHandler.GetQRImageHandler,
	getSessionInfoHandler *RouteHandler.GetSessionInfoHandler,
	getActiveConnectionInfoHandler *RouteHandler.GetActiveConnectionInfoHandler,
) error {
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
	router.Handle("/get-qr-code/{sessionId}/", getQRImageHandler).Methods("GET")
	router.Handle("/get-session-info/{sessionId}/", getSessionInfoHandler).Methods("GET")
	router.Handle("/get-active-connection-info/{sessionId}/", getActiveConnectionInfoHandler).Methods("GET")

	var err error
	certFileExists, certKeyExists := true, true
	if _, err = os.Stat(conf.CertFilePath); os.IsNotExist(err) {
		certFileExists = false
	}
	if _, err = os.Stat(conf.CertKeyPath); os.IsNotExist(err) {
		certKeyExists = false
	}

	if conf.Env == config.DevMode {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if !certFileExists || !certKeyExists {
		log.Printf("wapi will handle request by unsecured connection. Wapi's listening at %s ...\n", conf.ListenHttpHost)
		err = http.ListenAndServe(conf.ListenHttpHost, router)
	} else {
		log.Printf("wapi's listening at %s ...\n", conf.ListenHttpHost)
		err = http.ListenAndServeTLS(conf.ListenHttpHost, conf.CertFilePath, conf.CertKeyPath, router)
	}
	if err != nil {
		return err
	}
	return nil
}

func initRedis(redisHost string) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{Addr: redisHost})
	_, err := redisClient.Ping().Result()
	if err != nil {
		return nil, err
	}
	return redisClient, nil
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
