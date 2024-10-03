# Local Development

Change to the `local/` directory.

```
cd local
```

Output of `go run main.go`:

```
$ go run main.go
Targets:
  local:accessToken       prints the Microsoft Entra ID access token
  local:ip                prints the public IPv4 address of the current machine
  local:psql              <resourceGroup> opens a psql shell to the Postgres server using the current logged in user and the access token from az account get-access-token
  local:test              <name> prints a test message
  local:updateAdmin       <resourceGroup> updates the admin user of the Postgres server to the currently signed in user using postgres-admin.bicep
  local:updateFirewall    <resourceGroup> updates the firewall rule of the Postgres server to the current machine's IP address
```

## 1.1 Usage

For the Postgres server in `240900-linux-postgres`, update the firewall to the current machine's IP address, update the admin user to the current user, and connect to the server using `psql`.

```
go run main.go \
    local:updatefirewall 240900-linux-postgres \
    local:updateadmin 240900-linux-postgres \
    local:psql 240900-linux-postgres
```