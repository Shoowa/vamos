#!/bin/bash

curl -H "X-Vault-Token: token" \
  -H "Content-Type: application/json" \
  --request POST --data '{"data" : {"cert": "'"$(base64 --wrap=0 /data/ca/public/app.pem)"'"}}' \
  http://localhost:8200/v1/secret/data/dev-app-cert

curl -H "X-Vault-Token: token" \
  -H "Content-Type: application/json" \
  --request POST --data '{"data" : {"private_key": "'"$(base64 --wrap=0 /data/ca/private/app-key.pem)"'"}}' \
  http://localhost:8200/v1/secret/data/dev-app-key
