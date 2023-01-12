# HTTP to MQTT Bridge

/post to publish
/publish to publish
/subscribe to scribe

## Configuration

Running locally: [config.yaml](config.yaml) and `config.user.yaml` (gitignored, e.g. for credentials)

| Name             | Default                                     | Description                                      |
| ---------------- | ------------------------------------------- | ------------------------------------------------ |
| `logging.level`  | `INFO`                                      | The minimum log level to output                  |
| `logging.format` | `json`                                      | The format to use for log output (plain or json) |
| `port`           | `8080`                                      | The port to serve the gateway on                 |
