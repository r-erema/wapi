# wapi #

## Что это такое ##

**wapi** представляет собой api-обёртку над [**клиентом**](https://github.com/Rhymen/go-whatsapp), который подключается по вебсокетам к серверу WhatsApp. Более детально о работе [**клиента**](https://github.com/Rhymen/go-whatsapp) можно почитать [здесь](https://github.com/sigalor/whatsapp-web-reveng)

## Как это работает ##

Прежде всего надо зарегистрировать сессию, с помощью api-метода
>POST /register-session/  
>{"session_id": "%session_name_string%"}

`%session_name_string%` может быть произвольной строкой.  

В процессе регистрации проверяется существует ли файл сессии (.gob файл, будет лежать в `WAPI_FILE_SYSTEM_ROOT_POINT_FULL_PATH/sessions` ), если да, то будет произведена авторизация с помощью этого файла, иначе будет сгенерирован QR-код(он будет выведен в консоль и создастся файл с картинкой в `WAPI_FILE_SYSTEM_ROOT_POINT_FULL_PATH/qr-codes`) который надо сканировать через приложение WhatsApp в девайсе. После этого произойдёт авторизация и создастся файл сессии, также запустится слушатель который отправляется сообщения(адресованные на аккаунт WhatsApp с которого проходила авторизация) на вебхук `WAPI_GETTING_MESSAGES_WEBHOOK/%session_name_string%`


## Настройка ##
Существует несколько параметров представленных переменными окружения:
### Обязательные параметры ###
* **WAPI_GETTING_MESSAGES_WEBHOOK** - урл на который будут дублироваться сообщения отправленные с одного аккаунта WhatsApp другому, например:  `https://eggnog.chat/webhook/`   
* **WAPI_INTERNAL_HOST** - хост и порт который будет слушать веб-сервер api, например: `0.0.0.0:443`  
* **WAPI_FILE_SYSTEM_ROOT_POINT_FULL_PATH** - полный путь в файловой системе куда wapi будет складывать необходимые файлы(картинки qr-кодов, файлы сессии), например: `/home/user/wapi/files`  
* **WAPI_REDIS_HOST** - хост сервера Redis, например: `localhost:6379`  

### Необязательные параметры ### 
* **WAPI_SENTRY_DSN** - строка dsn для покдлючения к аккаунту Sentry, если не указан взаимодействия с Sentry не будет, например: `https://a58d45e33ec54df2802cf61c9651d123@sentry.io/9853030`  
* **WAPI_CERT_FILE_PATH** - путь к файлу сертификата, если не указан взаимодействие с api будет по http, например: `~/.ssl/cert.crt`  
* **WAPI_CERT_KEY_PATH** - путь к ключу сертификата, если не указан взаимодействие с api будет по http, например: `~/.ssl/cert.key`  
* **WAPI_ENV** - окружение wapi, допустимые значения `dev` или `prod`, если его значение `dev` то сертификат верифицироваться не будет  
* **WAPI_CONNECTIONS_CHECKOUT_DURATION_SECS** - интервал пинга соединений по вебсокетам всех зарегистрированных сессий, в секундах, по умолчанию `600`

## Методы api ##

* **Создание постоянного соединения по вебсокетам с сервером WhatsApp**  
>POST /register-session/  
`{
    "session_id": "%session_name_string%"
}`  

* **Отправка сообщения**  
> POST /send-message/  
`{  
    "chat_id":"375447034810@s.whatsapp.net",  
    "text":"test text",
    "session_name":"%session_name_string%"
}`  

* **Получение картинки QR кода**
> GET /get-qr-code/{sessionId}/  


* **Информация о конкретной сессии**  
> GET /get-session-info/{sessionId}/  

* **Информация о подключении по вебсокетам конкретной сессии**  
> GET /get-session-info/{sessionId}/  
