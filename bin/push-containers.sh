#!/bin/bash

source ./bin/env.sh

tag=$(git rev-parse HEAD)

echo Loggin into registry...
az acr login --name $REGISTRY

echo Building images...
docker build -t music-api:$tag .

echo Fetching login server info...
loginServer=$(
  az acr list --resource-group $GROUP \
    --query "[].{acrLoginServer:loginServer}" \
    --output tsv
)

echo Tagging images...
docker tag music-api:$tag $loginServer/music-api:$tag

echo Pusing images...
docker push $loginServer/music-api:$tag
