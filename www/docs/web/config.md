---
id: config
title: Configuration
---

`ghz-web` can be configured using environment variables or a configuration file.

## Environment Variables

- `GHZ_SERVER_PORT` - The port for the http server. Default is `80`.
- `GHZ_DATABASE_TYPE` - The SQL database dialect / type. Default is `sqlite3`.
- `GHZ_DATABASE_CONNECTION` - The SQL database connection string. Default is `data/ghz.db`.
- `GHZ_LOG_LEVEL` - The log level. One of `debug`, `info`, `warn`, or `error`. Default is `info`.
- `GHZ_LOG_PATH` - By default the logs go to `stdout`. This option can be used to set the log path for a log file.

## Configuration File

A cofiguration file can be specified using `-config` option. Configuration file can be in YAML, TOML or JSON format.

**YAML**

```yaml
---
server:
  port: 3000    # the port for the http server
database:       # the database options
  type: sqlite3
  connection: data/ghz.db
log:
  level: info
  path: /tmp/ghz.log # the path to log file, otherwize stdout is used
```

**TOML**

```toml
[server]
port = 80   # the port for the http server

[database]  # the database options
type = "sqlite3"
connection = "data/ghz.db"

[log]
level = "info"          # log level
path = "/tmp/ghz.log"   # the path to log file, otherwize stdout is used
```

**JSON**

```json
{
  "server": {
    "port": 80
  },
  "database": {
    "type": "sqlite3",
    "connection": "data/ghz.db"
  },
  "log": {
    "level": "info",
    "path": "/tmp/ghz.log"
  }
}

```

## Database

| Dialect  |                              Connection                              |
| :------: | :------------------------------------------------------------------: |
| sqlite3  | `path/to/database.db`                                                |
|  mysql   | `dbuser:dbpassword@/ghz`                                               |
| postgres | `host=dbhost user=dbuser dbname=ghz sslmode=disable password=dbpassword` |

When using postgres without SSL then `sslmode=disable` must be added to the connection string.
When using mysql with host then `tcp(host)` must be added to the connection string like that `dbuser:dbpassword@tcp(dbhost)/ghz`.
