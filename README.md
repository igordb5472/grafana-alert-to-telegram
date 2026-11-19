# grafana-alert-to-telegram (ga2tg)

Service that receives Grafana Alert JSON payloads (via webhook) and sends notifications to Telegram chats. You can create a custom message template and assign it to any chat or bot.

ga2tg also supports duty schedules with usernames to tag in duty notification. You can create a duty schedule and assign it to any chat or bot as well.

# Setup and run

## 1) Set up your Telegram bots:

### 1.1) Create Telegram bots which will be routing alerts from Grafana.

### 1.2) Add these bots to Telegram chats (groups/channels/...) where you want to receive the messages.

You must allow the bot to send messages. Ensure you've done it in every chat where you wait for messages from bot.

## 2) Set up and run your ga2tg instance:

### 2.1) Download one of ga2tg releases to your machine:

To download the latest one, proceed to the [latest release page](https://github.com/igordb5472/releases/latest/download) and download `.zip` or `.tar.gz` archive depending on which one is more convinient for you to extract. Extract the archive. After extracting you will have a strictly structured folder:

```
ga2tg-<version>/
├──bin/
├─────ga2tg-linux-amd64
├─────ga2tg-linux-arm64
├──settings/
├───── alert-notif-templates/
│      ├── alertNotif.tmpl
├───── bots/
│      ├── bot.json
├───── default-templates/
│      ├── alertNotif.default.tmpl
│      └── dutyNotif.default.tmpl
├───── duty-notif-templates/
│      ├── dutyNotif.tmpl
├───── duty-scheds/
│      ├── dutySched.json
├──.dockerignore
├──Dockerfile
└──example.env
```

Do not change this structure.

### 2.2) Configure environment variables:

Rename `example.env` to `.env` (or create a new `.env` file). Set up authorization variables (`AUTH_LOGIN` and `AUTH_PASSWORD`) and Telegram bot tokens in format `BOT<botID>_TOKEN=`:

```
AUTH_LOGIN=<login>
AUTH_PASSWORD=<password>
BOT123_TOKEN=<token>
BOT456_TOKEN=<token>
...
```

It's necessary to provide at least one bot id.

### 2.3) Configure Telegram bots and chats and message templates:

All configuration besides environment variables must be located in `settings/` folder with the following structure:

```
settings/
├── alert-notif-templates/
│   ├── alertNotif1.tmpl
│   ├── alertNotif2.tmpl
│   └── ... (*.tmpl)
├── bots/
│   ├── bot123.json
│   ├── bot456.json
│   └── ... (*.json)
├── default-templates/
│   ├── alertNotif.default.tmpl
│   └── dutyNotif.default.tmpl
├── duty-notif-templates/
│   ├── dutyNotif1.tmpl
│   ├── dutyNotif2.tmpl
│   └── ... (*.tmpl)
└── duty-scheds/
    ├── dutySched1.json
    ├── dutySched2.json
    └── ... (*.json)
```

a) In `bots/` set up your Telegram bots with their chats as JSON objects (one object ─ one bot) in `*.json` files with the following structure:

```
{
    "id": 123,
    "token": "<token>",
    "alertNotifTemplate": "alertNotif1.tmpl",
    "dutyNotifTemplate": "dutyNotif1.tmpl",
    "dutySched": "dutySched1.json",
    "chats": [
        {
            "id": -123,  // required
            "alertNotifTemplate": "alertNotif.tmpl",
            "dutyNotifTemplate": "dutyNotif.tmpl",
            "dutySched": "dutySched.json"
        },
        {
            "id": -456,
            ...
        }
        ...
    ]
}
```

Variables' reference:

(required) `id` ─ Telegram bot/chat ID.

(required) `token` ─ Telegram bot token (full, i.e. `<botID>:...`).

`alertNotifTemplate` ─ filename of template for notifications based on received Grafana alert payloads; default if not provided (chat's default is the bot's template).

`dutyNotifTemplate` ─ filename of template for duty notifications; default if not provided (chat's default is the bot's template).

`dutySched` ─ filename of duty schedule configuration (described later); absent (or the same as bots for chats) if not provided.

All bots must be configured separately (one bot ─ one file); it's necessary to provide at least one bot configuration file.

b) In `duty-scheds/` set up your duty schedules as JSON arrays of JSON objects, each one of which represents a duty notification, in `*.json` files with the following structure:

```
[
    {
        "date": "dd.mm.yyyy",
        "time": "hh:mm",
        "timezone": "UTC+hh:mm",
        "repeat": "<number><unit>",
        "users": [
            "@username1",
            "@username2",
            ...
        ]
    },
    {
        "date": ...,
        ...
    },
    ...
]
```

Variables' reference:

(required) `date` ─ notification date.

(required) `users` ─ array of Telegram usernames to tag; must be non-empty with proper usernames in format `@...`; you can access an array of usernames using `.usernames` in your template.

`time` ─ notification time; 00:00 if not provided.

`timezone` ─ timezone of the notification time; UTC+00:00 if not provided; UTC-hh:mm format is also supported.

`repeat` ─ interval at which the messages are repeated; notification is sent once if not provided (or never if configured in the past); supported units of time measurement: m ─ minute, h ─ hour, d ─ day, w ─ week; so you can provide "1h", "3d", "10m" and so on.


c) In `alert-notif-templates/` set up templates for notifications based on received Grafana alert payloads, in `duty-notif-templates` ─ for duty notifications. They must be written in Go template language and have `.tmpl` extension. You can provide any valid Go template. If the resulting message is empty, it won't be sent by the bot which the template is applied to to its chats. You can use this behaviour for giving permissions to a bot in your template (e.g. `{{if .status == "firing"}}...`).


d) In `default-templates/` set up default template for notifications based on received Grafana alert payloads and default template for duty notifications. You absolutely must not change the structure of this folder: add, rename or remove templates. You should only change the code of these 2 templates if you want to.

### 2.4) Build and run your image:

Run the following command from `ga2tg/`:

```bash
docker buildx build -t ga2tg . && docker run --rm -p 8080:8080 --add-host=host.docker.internal:host-gateway ga2tg
```

## 3) Configure your Grafana webhook

### 3.1) Add your ga2tg instance URL.

ga2tg is listening on `host.docker.internal:8080`.

### 3.2) Add Basic Authentication Username and Basic Authentication Password.

They must match AUTH_LOGIN and AUTH_PASSWORD from your .env respectively.
