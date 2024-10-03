# ENV

```bash
PG_NAME=$(az postgres flexible-server list \
    --resource-group 240900-linux-postgres \
    --query "[0].name" \
    -o tsv)

# USER_NAME=$(az account show --query 'user.name' -o tsv)

USER_NAME='aawislan_microsoft.com#EXT#@cacloudnative2411.onmicrosoft.com'

export PGHOST="${PG_NAME}.postgres.database.azure.com"
export PGPASSWORD=$(go run main.go app:token)
export PGUSER=$USER_NAME
export PGDATABASE=postgres
```