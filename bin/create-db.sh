#!/bin/bash

source ./bin/env.sh

echo Creating resource group...
az group create --name $GROUP --location $LOCATION

echo Creating postgres server...
az postgres server create \
  --location $LOCATION \
  --resource-group $GROUP \
  --name $DB_SERVER_NAME \
  --admin-user $ADMINUSER \
  --admin-password $ADMINPWORD \
  --sku-name GP_Gen5_2

echo Configuring server firewall rule...
az postgres server firewall-rule create \
  --resource-group $GROUP \
  --server $DB_SERVER_NAME \
  --name AllowAllIps \
  --start-ip-address 0.0.0.0 \
  --end-ip-address 255.255.255.255

export PGPASSWORD=$ADMINPWORD

psql -U $ADMINUSER@$DB_SERVER_NAME -h $DB_SERVER_NAME.postgres.database.azure.com \
  -c "CREATE DATABASE $DB_DATABASE;" postgres

psql -U $ADMINUSER@$DB_SERVER_NAME -h $DB_SERVER_NAME.postgres.database.azure.com \
  -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" postgres

psql -U $ADMINUSER@$DB_SERVER_NAME -h $DB_SERVER_NAME.postgres.database.azure.com \
  -c "GRANT ALL PRIVILEGES ON DATABASE $DB_DATABASE TO $DB_USER;" postgres
