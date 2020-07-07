package main

import (
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/Rhymen/go-whatsapp"
	"github.com/getsentry/sentry-go"
	"github.com/r-erema/wapi/internal/config"
	httpInternal "github.com/r-erema/wapi/internal/http"
	"github.com/r-erema/wapi/internal/repository"
	messageRepo "github.com/r-erema/wapi/internal/repository/message"
	sessionRepo "github.com/r-erema/wapi/internal/repository/session"
	"github.com/r-erema/wapi/internal/service"
)

func main() {
	conf, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	initSentry(conf)

	msgRepo := msgRepo(conf)
	sessRepo := sessRepo(conf)
	connSupervisor := connSupervisor(conf)
	resolver := qrFileResolver(conf)
	authorizer := authorizer(conf, sessRepo, connSupervisor, resolver)
	listener := service.NewWebHook(sessRepo, connSupervisor, authorizer, conf.WebHookURL, msgRepo, &http.Client{})

	router := httpInternal.Router(conf, sessRepo, connSupervisor, authorizer, resolver, listener)

	certFileExists, certKeyExists := true, true
	if _, err = os.Stat(conf.CertFilePath); os.IsNotExist(err) {
		certFileExists = false
	}
	if _, err = os.Stat(conf.CertKeyPath); os.IsNotExist(err) {
		certKeyExists = false
	}

	if !certFileExists || !certKeyExists {
		log.Printf("wapi will handle request by unsecured connection. Wapi's listening at %s ...\n", conf.ListenHTTPHost)
		err = http.ListenAndServe(conf.ListenHTTPHost, router)
	} else {
		log.Printf("wapi's listening at %s ...\n", conf.ListenHTTPHost)
		err = http.ListenAndServeTLS(conf.ListenHTTPHost, conf.CertFilePath, conf.CertKeyPath, router)
	}
	log.Fatal(err)
}

func msgRepo(conf *config.Config) repository.MessageRepository {
	log.Print("connecting to redis...")
	msgRepo, err := messageRepo.NewRedis(conf.RedisHost)
	if err != nil {
		log.Fatalf("error of init redis repo: %v\n", err)
	}
	log.Print("ok")
	return msgRepo
}

func sessRepo(conf *config.Config) repository.SessionRepository {
	sessRepo, err := sessionRepo.NewFileSystem(conf.FileSystemRootPath + "/sessions")
	if err != nil {
		log.Fatalf("can't create service `session`: %v\n", err)
	}
	return sessRepo
}

func connSupervisor(conf *config.Config) service.Connections {
	return service.NewSV(time.Duration(conf.ConnectionsCheckoutDuration))
}

func qrFileResolver(conf *config.Config) service.QRFileResolver {
	qrFileResolver, err := service.NewQRImgResolver(conf.FileSystemRootPath + "/qr-codes")
	if err != nil {
		log.Fatalf("can't create service `QR file resolver`: %v\n", err)
	}
	return qrFileResolver
}

func authorizer(
	conf *config.Config,
	sessRepo repository.SessionRepository,
	connSupervisor service.Connections,
	resolver service.QRFileResolver,
) service.Authorizer {
	authorizer := service.New(
		time.Duration(conf.ConnectionTimeout)*time.Second,
		sessRepo,
		connSupervisor,
		resolver,
	)
	return authorizer
}

func initSentry(conf *config.Config) {
	if conf.SentryDSN != "" {
		log.Print("init Sentry...")
		if err := sentry.Init(sentry.ClientOptions{Dsn: conf.SentryDSN}); err != nil {
			log.Fatalf("error of init sentry: %v\n", err)
		}
		log.Print("ok")
	} else {
		log.Print("sentry dsn not set, skip sentry init")
	}
}
