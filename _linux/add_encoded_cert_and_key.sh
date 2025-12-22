#!/bin/bash

curl -H "X-Vault-Token: token" \
  http://localhost:8200/v1/secret/data/dev-app-cert \
  -H "Content-Type: application/json" \
  --request POST \
  --data @- <<EOF
{
  "data" :
  {
    "cert": "$(base64 --wrap=0 /data/ca/public/app.pem)",
    "client_cert": "$(base64 --wrap=0 /data/ca/public/app_client.pem)"
  }
}
EOF

curl -H "X-Vault-Token: token" \
  http://localhost:8200/v1/secret/data/dev-app-key \
  -H "Content-Type: application/json" \
  --request POST \
  --data @- <<EOF
{
  "data" :
  {
    "private_key": "$(base64 --wrap=0 /data/ca/private/app-key.pem)",
    "client_private_key": "$(base64 --wrap=0 /data/ca/private/app_client-key.pem)"
  }
}
EOF

curl -H "X-Vault-Token: token" \
  http://localhost:8200/v1/secret/data/intermediate-ca \
  -H "Content-Type: application/json" \
  --request POST \
  --data @- <<EOF
{
  "data" :
  {
    "int_ca": "$(base64 --wrap=0 /data/ca/root/ca.pem)"
  }
}
EOF
