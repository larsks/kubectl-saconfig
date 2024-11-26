#!/bin/sh

name="$1"

if [ -z "$name" ]; then
  echo "ERROR: missing service account name" >&2
  exit 1
fi

token=$(kubectl create token "$name")
if [ -z "$token" ]; then
  echo "EROR: failed to acquire token for $name" >&2
  exit 1
fi

kubectl config view --flatten --minify -o json | jq \
  --arg service_account_name "sa-$name" \
  --arg service_account_token "$token" \
  '
.contexts as $CONTEXTS |
.contexts = [
  {
    "name": $service_account_name,
    "context": {
      "cluster": $CONTEXTS[0].context.cluster,
      "user": $service_account_name,
      "namespace": $CONTEXTS[0].context.namespace
    }
  }
] |
."current-context" = $service_account_name |
.users = [
  {
    "name": $service_account_name,
    "user": {
      "token": $service_account_token
    }
  }
]
'
