# End-to-end tests

Two suites, run by `../e2e-test.sh` against the docker environment from
`../init-tests-env.sh`. They answer different questions and are deliberately kept apart.

Both use a provider built from the working tree, wired in through a generated
`.terraformrc` with `dev_overrides`.

## `surfaceTests/` — "can this be expressed?"

A single wide, flat config: a namespace with every optional field set, fake `s3` and
`aws-secret-manager` blocks, a nested namespace, a second tenant, and real-world flows
loaded through `for_each` + `templatefile` + `yamldecode`.

It applies, then requires the follow-up plan to be empty. It makes **no behavioural
assertions, on purpose** — its value is breadth of schema surface, and several resources
here are exercised nowhere else inside a real dependency graph. Its oddness is a feature;
please don't tidy it into something coherent.

## `scenarioTests/` — "does it actually work?"

A customer bootstrap in stages, each running as a **different Kestra identity**. A stage
can only run once the previous one has created the credentials it authenticates with,
which is what keeps the provider from being tested exclusively as super-admin.

| Stage | Runs as | Builds |
|---|---|---|
| `00-bootstrap/` | the super-admin from `kestra.security.super-admin` | the platform admin: user, password, role, tenant-wide binding, API token |
| `10-platform/` | the platform admin created above | namespaces, secret, KV, namespace file, flows, test suite, the app user and its narrow group/role/binding |
| `runtime/` | the app user (and the platform admin) | no Terraform — Go assertions over the API |

The super-admin is an **input**, never a managed resource: Kestra has no API for creating
the first one, and `kestra_user` has never exposed the privilege. Stage 00 is the only
stage that uses those break-glass credentials, which is also the only stage a customer
would run with them.

### What `runtime/` asserts

- the app user can run the flow, it reaches `SUCCESS`, and its output matches the fixtures
  — which also proves the secret, KV entry and namespace file resolve at runtime
- the app user **cannot** create flows, and cannot read or run a flow in a sibling
  namespace its binding does not cover
- the app user **can** read the flow it is entitled to (so the suite cannot pass with a
  role that grants nothing)
- the `kestra_test` suite Terraform manages actually runs clean

Test files are behind the `e2e` build tag, so ordinary `go build`/`go vet`/`go test` runs
never compile them.

### Namespaces

`io.kestra.e2escenario.allowed` and `io.kestra.e2escenario.forbidden` are **siblings**.
Kestra extends a namespace binding down to child namespaces, so a nested "forbidden"
namespace would inherit access and the denial assertions would pass for the wrong reason.
They are also disjoint from the acceptance-test fixtures (`io.kestra.terraform.*`) and
from `surfaceTests` (`io.kestra.terraform.e2e.*`).

### Role actions

Use the Kestra 2.0 fine-grained action names (`VIEW`, `LIST`, `EXECUTE`, `ACCESS_LOGS`, …),
not the pre-2.0 CRUD verbs. The old `READ`/`CREATE`/`UPDATE`/`DELETE` names still
round-trip through the API, so a role written with them looks right in state and in a plan
while granting nothing. `internal/provider/migrate_role_permissions.go` has the full
mapping.

## Not yet covered

- `surfaceTests` has no `destroy` phase, so delete ordering across its graph is untested.
- No drift assertion yet: `bodyToNamespaceModel` keeps prior values for absent optional
  keys, so a namespace setting cleared outside Terraform produces no plan.
- The scenario runs in the `main` tenant. A dedicated tenant would isolate it from the
  acceptance fixtures.
