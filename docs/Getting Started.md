# How to run

Create a `config` folder at the root directory of the project and inside that you can put configuration files for running in different environments.
The application by default reads the configuration from `config/production.yml`. You can pass a different filename by including the `-config` option when running the application. For example:

```
$ go run main.go -config=debug.yml
```

## Configuration file

You need to define database connection information in the configuration file in `yaml` format.
example configuration file (`config/debug.yml`):

```
dsn: "<database_user>:<database_password>@tcp(localhost:3306)/<database_name>?parseTime=true"
```
