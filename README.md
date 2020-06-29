[![Build Status](https://travis-ci.com/r-erema/wapi.svg?branch=master)](https://travis-ci.com/r-erema/wapi)

# wapi #

## What is it? ##

**Wapi** is an api-wrapper over [**client**](https://github.com/Rhymen/go-whatsapp), that connects via web sockets to the WhatsApp server. More about client work [here](https://github.com/sigalor/whatsapp-web-reveng).

## How it works ##

First of all, you need to register a session using the api method:
>POST /register-session/  
>{"session_id": "%session_name_string%"}

`%session_name_string%` may be an arbitrary string.  

During the registration process, it is checked whether the session file exists (.gob file, locates in `WAPI_FILE_SYSTEM_ROOT_POINT_FULL_PATH/sessions` ), if yes authorization will be performed using this file, otherwise a QR code will be generated(it will be outputed in the console and file with a picture will be created in  `WAPI_FILE_SYSTEM_ROOT_POINT_FULL_PATH/qr-codes`) which must be scanned by the WhatsApp application in the device(e.g. smartphone). After that, authorization will occur and session file will be created, a listener will also be launched that sends messages (addressed to the WhatsApp account from which the authorization took place) on the webhook `WAPI_GETTING_MESSAGES_WEBHOOK/%session_name_string%`


## Settings ##
There are several parameters represented by environment variables:
### Required parameters ###
* **WAPI_GETTING_MESSAGES_WEBHOOK** - URL to which messages sent from one WhatsApp account to another will be duplicated, e.g. `https://eggnog.chat/webhook/`   
* **WAPI_INTERNAL_HOST** - host and port, that the api web server will listen to, e.g. `0.0.0.0:443`  
* **WAPI_FILE_SYSTEM_ROOT_POINT_FULL_PATH** - full path in the file system where wapi stores the necessary files (pictures of qr codes, session files), e.g. `/home/user/wapi/files`  
* **WAPI_REDIS_HOST** - Redis server host, e.g. `localhost:6379`  

### Optional parameters ### 
* **WAPI_SENTRY_DSN** - dsn line to connect to the Sentry account, e.g. `https://a58d45e33ec54df2802cf61c9651d123@sentry.io/9853030`. If it is not specified there will be no interaction with Sentry
* **WAPI_CERT_FILE_PATH** - the path to the certificate file, e.g. `~/.ssl/cert.crt`  
* **WAPI_CERT_KEY_PATH** - path to certificate key, e.g. `~/.ssl/cert.key`  
* **WAPI_ENV** - wapi environment, valid `dev` or` prod` values, if its value is `dev`, then the certificate will not be verified  
* **WAPI_CONNECTIONS_CHECKOUT_DURATION_SECS** - interval of ping connections on web sockets of all registered sessions, in seconds, by default `600`

## Api methods ##

* **Creating a web socket connection to a WhatsApp server**  
>POST /register-session/  
`{
    "session_id": "%session_name_string%"
}`  

* **Message sending**  
> POST /send-message/  
`{  
    "chat_id":"375447034810@s.whatsapp.net",  
    "text":"test text",
    "session_name":"%session_name_string%"
}`  

* **Getting a picture of a QR code**
> GET /get-qr-code/{sessionId}/  


* **Session information**  
> GET /get-session-info/{sessionId}/  

* **Web Socket connection information of particular session**  
> GET /get-session-info/{sessionId}/  
