# tt-go

tt-go is an application workload, written in Go, that targets Linux and Postgres. It demonstrates how to:

- Access the Postgres server locally or from a VM using Entra ID and a Managed Identity (User Assigned) using Go and the
`azidentity` package.
- Optionally configure the Firewall and Administrator for the Postgres server and access it locally using psql and Entra ID for your own user.

This project uses [Mage](https://magefile.org/), a make-like build tool written in Go. Mage enables us to write `targets` as simple Go functions. If you have mage installed, you can run `mage` instead of `go run main.go`. However, we include mage as a library via the [zero install](https://magefile.org/zeroinstall/) so you can run `go run main.go` without installing mage.

These workflows are designed to be both developer and automation-friendly, including GitHub Actions and Executable Docs.

## Requirements
- [Go (1.22+)](https://go.dev/doc/install)
- [psql](https://www.postgresql.org/docs/current/app-psql.html) CLI (e.g. `apt-get install postgresql-client` or `brew install libpq`)
- (Optional) [Mage](https://magefile.org/)


## Application

Change to the `app/` directory.

```
cd app
```

Output of `go run main.go`:

```
$ go run main.go
Targets:
  app:connectionString    outputs a connection string for the database from env vars
  app:ping                pings the database
  app:serve               runs a web server for our application
  app:tables              lists the tables in the database
  app:token               gets a token using `azidentity.NewDefaultAzureCredential`
```

## Local Development

See the [local/README.md](local/README.md) for more information.
