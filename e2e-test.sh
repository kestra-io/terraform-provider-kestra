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
echo ""
echo "terraform plan"
exit_code=0
TF_CLI_CONFIG_FILE=$(pwd)/e2e-tests/.terraformrc TF_LOG_PATH=terraform-trace.log TF_LOG=DEBUG terraform -chdir=./e2e-tests plan -detailed-exitcode || exit_code=$?
echo ""
echo "checking terraform plan is empty"
if [ $exit_code -eq 0 ]; then
  echo "✅ No changes to apply."
elif [ $exit_code -eq 2 ]; then
  echo "⚠️  There are changes to apply. It should not happen after an apply, something is wrong in the provider for these specific Resources"
  exit 2
else
  echo "❌ Terraform plan failed."
  exit 1
fi
echo ""
echo "this test has no assertions yet, but playing apply and plan we check that: providers/mux config is working, a real world example is working"
