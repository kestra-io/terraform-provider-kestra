<p align="center">
  <a href="https://www.kestra.io">
    <img src="https://kestra.io/banner.png"  alt="Kestra workflow orchestrator" />
  </a>
</p>

<h1 align="center" style="border-bottom: none">
    Event-Driven Declarative Orchestrator
</h1>

<div align="center">
 <a href="https://github.com/kestra-io/kestra/releases"><img src="https://img.shields.io/github/tag-pre/kestra-io/kestra.svg?color=blueviolet" alt="Last Version" /></a>
  <a href="https://github.com/kestra-io/kestra/blob/develop/LICENSE"><img src="https://img.shields.io/github/license/kestra-io/kestra?color=blueviolet" alt="License" /></a>
  <a href="https://github.com/kestra-io/kestra/stargazers"><img src="https://img.shields.io/github/stars/kestra-io/kestra?color=blueviolet&logo=github" alt="Github star" /></a> <br>
<a href="https://kestra.io"><img src="https://img.shields.io/badge/Website-kestra.io-192A4E?color=blueviolet" alt="Kestra infinitely scalable orchestration and scheduling platform"></a>
<a href="https://kestra.io/slack"><img src="https://img.shields.io/badge/Slack-Join%20Community-blueviolet?logo=slack" alt="Slack"></a>
</div>

<br />

<p align="center">
    <a href="https://twitter.com/kestra_io"><img height="25" src="https://kestra.io/twitter.svg" alt="twitter" /></a> &nbsp;
    <a href="https://www.linkedin.com/company/kestra/"><img height="25" src="https://kestra.io/linkedin.svg" alt="linkedin" /></a> &nbsp;
<a href="https://www.youtube.com/@kestra-io"><img height="25" src="https://kestra.io/youtube.svg" alt="youtube" /></a> &nbsp;
</p>

<br />
<p align="center">
    <a href="https://go.kestra.io/video/product-overview" target="_blank">
        <img src="https://kestra.io/startvideo.png" alt="Get started in 4 minutes with Kestra" width="640px" />
    </a>
</p>
<p align="center" style="color:grey;"><i>Get started with Kestra in 4 minutes.</i></p>


# Kestra Terraform Provider

This repository defines Kestra resources so that they can be deployed using Infrastructure as Code with Terraform.

> [!IMPORTANT]  
> Kestra Terraform provider 0.23.x is only compatible with Kestra 0.23.x and above.
> Additionally, if you want to terraform Kestra 0.23.x you need to use Kestra Terraform provider 0.23.x

## Documentation

* The official Kestra documentation can be found under [kestra.io/docs](https://kestra.io/docs)
* Kestra Terraform provider documentation can be found [here](https://kestra.io/docs/terraform/).


## Using the provider

This Terraform Provider is available to install automatically via `terraform init`. It is recommended to setup the following Terraform configuration to pin the major version:

```hcl
terraform {
  required_providers  {
    source = "kestra-io/kestra"
    version = "~> X.Y" # where X.Y is the current major version and minor version
  }
}
```

Additional documentation, including available resources and their arguments/attributes can be found on the [Terraform documentation website](https://registry.terraform.io/providers/kestra-io/kestra/latest/docs).

## Developing the Provider
If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and need a Kestra cluster.

```sh
$ make testacc
```

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
- [Go](https://golang.org/doc/install) >= 1.16

### Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:
```sh
$ go install
```

### Start tests in local
The full test suite requires to start the full [docker-compose-ci.yml](docker-compose-ci.yml) and have access to Kestra EE docker image:

1. read and do requirements of [init-tests-env.sh](init-tests-env.sh)
2. init test environment 
```sh 
$ ./init-tests-env.sh
```
3. run tests
```sh
$ TF_ACC=1 KESTRA_URL=http://127.0.0.1:8088 KESTRA_USERNAME=root@root.com KESTRA_PASSWORD='Root!1234' go test -v -cover ./internal/provider/
```
#### Test coverage
To display generate a test coverage file in local you can add `-coverprofile` to your test command:
```
go test -v -coverprofile=test-coverage-result.out ./internal/provider/
```
and then to display it:
```
# in browser
$ go tool cover -html=test-coverage-result.out

# in terminal
$ go tool cover -func=test-coverage-result.out
```
### Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.


## Stay up to date

We release new versions every month. Give the [main repository](https://github.com/kestra-io/kestra) a star to stay up to date with the latest releases and get notified about future updates.

![Star the repo](https://kestra.io/star.gif)

## License
Apache 2.0 © [Kestra Technologies](https://kestra.io)
