#!/bin/bash

source ./bin/env.sh

tag=$(git rev-parse HEAD)

echo Fetching credentials...
password=$(
  az acr credential show \
    --name $REGISTRY \
    --query "passwords[0].value" \
    --output tsv
)

echo Creating App Service plan...
az appservice plan create \
  --name $PLAN \
  --resource-group $GROUP \
  --location $LOCATION \
  --sku B1 \
  --is-linux

echo Creating web app...
az webapp create \
  --resource-group $GROUP \
  --plan $PLAN \
  --name $APP \
  --deployment-container-image-name $REGISTRY.azurecr.io/music-api:$tag

echo Configuring web app...
az webapp config container set \
  --name $APP \
  --resource-group $GROUP \
  --docker-custom-image-name $REGISTRY.azurecr.io/music-api:$tag \
  --docker-registry-server-url https://$REGISTRY.azurecr.io \
  --docker-registry-server-user $REGISTRY \
  --docker-registry-server-password $password

echo Setting port...
az webapp config appsettings set \
  --resource-group $GROUP \
  --name $APP \
  --settings WEBSITES_PORT=8080
