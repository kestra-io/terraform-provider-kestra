set -e;

echo ""
echo "e2e tests, we use the existing docker env from ./init-tests-env.sh and we do a real terraform apply"
echo ""

echo ""
echo "go build"
echo ""
rm -f terraform-trace.log
rm -f ./e2e-tests/terraform.tfstate*

# hope it will work for all local dev setup
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
go build -o ./e2e-tests/.terraform/plugins/local/kestra-io/kestra/terraform-provider-kestra_0.23.0
echo ""
echo "terraform apply with the provider we just built"
echo ""
TF_CLI_CONFIG_FILE=$(pwd)/e2e-tests/.terraformrc TF_LOG_PATH=terraform-trace.log TF_LOG=DEBUG terraform -chdir=./e2e-tests apply -auto-approve
echo ""
echo "terraform apply succeded"
echo "this test has no assertions yet, but playing apply checks: providers/mux config is working, a real world example is working"
