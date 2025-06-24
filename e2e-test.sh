set -e;

echo ""
echo "e2e tests, we use the existing docker env from ./init-tests-env.sh and we do a real terraform apply"
echo ""
echo "terraform init"
echo ""
terraform -chdir=./e2e-tests init

echo ""
echo "go build"
echo ""
# hope it will work for all local dev setup
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
rm terraform-trace.log & go build -o .terraform/plugins/local/kestra-io/kestra/terraform-provider-kestra_0.23.0
echo ""
echo "terraform apply with the provider we just built"
echo ""
TF_LOG_PATH=terraform-trace.log TF_LOG=DEBUG terraform -chdir=./e2e-tests apply -auto-approve
echo ""
echo "terraform apply succeded"
echo "this test has no assertions yet, but playing apply checks: providers/mux config is working, a real world example is working"
