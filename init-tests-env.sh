#!/bin/bash
#===============================================================================
# SCRIPT: init-tests-env.sh
#
# DESCRIPTION:
#   Start and prepare a local Kestra EE + ES env for testing
#
# PREREQUISITE:
# the Kestra EE licence configuration in .github/docker/application-secrets.yml:
#     kestra:
#       ee:
#         license:
#           id: 5dcscsd-dfsdf.........c3
#           key: |
#             Ic5OXgAAAB............RVI=
#
# USAGE: ./ init-tests-env.sh
# then you can launch the tests with:
# TF_ACC=1 KESTRA_URL=http://127.0.0.1:8088 KESTRA_USERNAME=root@root.com KESTRA_PASSWORD='Root!1234' go test -v -cover ./internal/provider/
#
#===============================================================================

set -e;

echo "initializing test environment with docker compose"

docker compose -f docker-compose-ci.yml up elasticsearch kafka vault -d --wait || {
   echo "db Docker Compose failed. Dumping logs:";
   docker-compose -f docker-compose-ci.yml logs;
   exit 1;
}
docker compose -f docker-compose-ci.yml up kestra -d --wait || {
   echo "kestra Docker Compose failed. Dumping logs:";
   docker compose -f docker-compose-ci.yml logs kestra;
   exit 1;
}
sleep 10
docker compose -f docker-compose-ci.yml logs kestra;

echo "\echo 'path \"*\" {capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\", \"sudo\"]}' | vault policy write admins -" | docker exec --interactive terraform-provider-kestra-vault-1 sh -
docker compose -f docker-compose-ci.yml exec vault vault auth enable userpass
docker compose -f docker-compose-ci.yml exec vault vault write auth/userpass/users/john \
    password=foo \
    policies=admins \
    token_period=1s


curl --fail-with-body "127.27.27.27:9200"
#curl --fail-with-body -u 'root@root.com:Root!1234' -X POST "127.27.27.27:8080/api/v1/main/users"
curl --fail-with-body -u 'root@root.com:Root!1234' "127.27.27.27:8080/api/v1/main/users/search" > /dev/null
curl --fail-with-body -u 'root@root.com:Root!1234' -X POST -H 'Content-Type: application/json' -d '{"id":"unit_test","name":"Unit Test"}' "127.27.27.27:8080/api/v1/tenants" > /dev/null
curl --fail-with-body -H "Content-Type: application/x-ndjson" -XPOST "127.27.27.27:9200/_bulk?pretty" --data-binary @.github/workflows/index.jsonl > /dev/null
sleep 5

curl  --fail-with-body -H "Content-Type: application/json" -u 'root@root.com:Root!1234' -X POST -d '{"id":"io.kestra.terraform.data", "tenantId":"main", "description": "My Kestra Namespace data"}' "127.27.27.27:8080/api/v1/main/namespaces" > /dev/null

curl  --fail-with-body -H "Content-Type: multipart/form-data" -u 'root@root.com:Root!1234' -X POST -F "fileContent=@internal/resources/flow.py" "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/files?path=/flow.py" > /dev/null
curl  --fail-with-body -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d '"stringValue"' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/string"
curl  --fail-with-body -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d '1' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/int"
curl  --fail-with-body -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d '1.5' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/double"
curl  --fail-with-body -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d 'false' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/falseBoolean"
curl  --fail-with-body -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d 'true' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/trueBoolean"
curl  --fail-with-body -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d '2022-05-01T03:02:01Z' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/dateTime"
curl  --fail-with-body -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d '2022-05-01' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/date"
curl  --fail-with-body -H "Content-Type: text/plain" -u 'root@root.com:Root!1234' -X PUT -d 'P3DT3H2M1S' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/duration"
curl  --fail-with-body -H "Content-Type: application/json" -u 'root@root.com:Root!1234' -X PUT -d '{"some":"value","in":"object"}' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/object"
curl  --fail-with-body -H "Content-Type: application/json" -u 'root@root.com:Root!1234' -X PUT -d '[{"some":"value","in":"object"},{"yet":"another","array":"object"}]' "127.27.27.27:8080/api/v1/main/namespaces/io.kestra.terraform.data/kv/array"
