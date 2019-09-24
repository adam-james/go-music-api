#!/bin/bash

source ./bin/env.sh

echo Creating group...
az group create --name $GROUP --location $LOCATION

echo Creating registry...
az acr create \
  --name $REGISTRY \
  --resource-group $GROUP \
  --location $LOCATION \
  --sku Basic \
  --admin-enabled true
