#!/bin/bash
#===============================================================================
# SCRIPT: init-tests-env.sh
#
# DESCRIPTION:
#   Start and prepare a local Kestra EE + ES env for testing
#
# PREREQUISITES/REQUIREMENTS:
# the Kestra EE licence configuration in .github/docker/application-secrets.yml:
#     kestra:
#       ee:
#         license:
#           id: 5dcscsd-dfsdf.........c3
#           key: |
#             Ic5OXgAAAB............RVI=
#
# in local development you may not have access to the docker image, what you can do:
# change docker-compose-ci.yml image to europe-west1-docker.pkg.dev/kestra-host/docker/kestra-ee:develop
# log to GCP registry: echo $(gh auth token) | docker login ghcr.io -u $(gh api user --jq .login) --password-stdin
#
#
# USAGE: ./init-tests-env.sh
# then you can launch the tests with:
# TF_ACC=1 KESTRA_URL=http://127.0.0.1:8088 KESTRA_USERNAME=root@root.com KESTRA_PASSWORD='Root!1234' go test -v -cover ./internal/provider/
#
#===============================================================================

set -e;

echo "starting init-tests-env.sh"
echo ""
echo "docker compose down --volumes, need fresh databases"
docker compose -f docker-compose-ci.yml down --volumes --remove-orphans

echo ""
echo "--------------------------------------------"
echo ""
echo "initializing test environment with docker compose"
docker compose -f docker-compose-ci.yml up elasticsearch kafka vault -d --wait || {
   echo "db Docker Compose failed. Dumping logs:";
   docker compose -f docker-compose-ci.yml logs;
   exit 1;
}
echo ""
echo "--------------------------------------------"
echo ""
echo "start Kestra"
docker compose -f docker-compose-ci.yml up kestra -d --wait || {
   echo "kestra Docker Compose failed. Dumping logs:";
   docker compose -f docker-compose-ci.yml logs kestra;
   exit 1;
}
# all of these sleep are maybe not needed anymore
sleep 10

docker compose -f docker-compose-ci.yml logs kestra;

echo ""
echo "--------------------------------------------"
echo ""
echo "inject test data in Vault"
echo "echo 'path \"*\" {capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\", \"sudo\"]}' | vault policy write admins -" | docker exec --interactive terraform-provider-kestra-vault-1 sh -
docker compose -f docker-compose-ci.yml exec vault vault auth enable userpass
docker compose -f docker-compose-ci.yml exec vault vault write auth/userpass/users/john \
    password=foo \
    policies=admins \
    token_period=1s

echo ""
echo "--------------------------------------------"
echo ""
echo "do some basic healthchecks"
curl --fail-with-body -sS "127.27.27.27:9200"
curl --fail-with-body -sS -u 'root@root.com:Root!1234' "127.27.27.27:8080/api/v1/main/flows/search"

echo ""
echo "--------------------------------------------"
echo ""
echo "create unit_test tenant using Kestra API"
curl --fail-with-body -sS -u 'root@root.com:Root!1234' -X POST -H 'Content-Type: application/json' -d '{"id":"unit_test","name":"Unit Test"}' "127.27.27.27:8080/api/v1/tenants"

echo ""
echo "--------------------------------------------"
echo ""
echo "inject most test data directly into ES, yes this is ugly, we should migrate this to using Kestra API or better the new SDK"
curl --fail-with-body -sS -H "Content-Type: application/x-ndjson" -XPOST "127.27.27.27:9200/_bulk?pretty" --data-binary @.github/workflows/index.jsonl 
sleep 5

echo ""
echo "--------------------------------------------"
echo ""
echo "inject more test data using Kestra API (namespace, flow.py, kv data)"
curl  --fail-with-body -sS -H "Content-Type: application/json" -u 'root@root.com:Root!1234' -X POST -d '{"id":"io.kestra.terraform.data", "tenantId":"main", "description": "My Kestra Namespace data"}' "127.27.27.27:8080/api/v1/main/namespaces"
curl  --fail-with-body -sS -H "Content-Type: multipart/form-data" -u 'root@root.com:Root!1234' -X POST -F "fileContent=@internal/resources/flow.py" "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/files?path=/flow.py"
curl  --fail-with-body -sS -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d '"stringValue"' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/string"
curl  --fail-with-body -sS -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d '1' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/int"
curl  --fail-with-body -sS -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d '1.5' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/double"
curl  --fail-with-body -sS -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d 'false' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/falseBoolean"
curl  --fail-with-body -sS -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d 'true' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/trueBoolean"
curl  --fail-with-body -sS -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d '2022-05-01T03:02:01Z' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/dateTime"
curl  --fail-with-body -sS -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d '2022-05-01' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/date"
curl  --fail-with-body -sS -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d 'P3DT3H2M1S' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/duration"
curl  --fail-with-body -sS -H "Content-Type: application/json" -u 'root@root.com:Root!1234' -X PUT -d '{"some":"value","in":"object"}' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/object"
curl  --fail-with-body -sS -H "Content-Type: application/json" -u 'root@root.com:Root!1234' -X PUT -d '[{"some":"value","in":"object"},{"yet":"another","array":"object"}]' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/array"

echo ""
echo ""
echo "init-tests-env.sh finished successfully"
echo ""
