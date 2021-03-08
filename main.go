package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/metrics"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/open-policy-agent/opa/util"
	"github.com/tsandall/opa-persistent-store-exp/persistent"
)

var pump = flag.Bool("pump", false, "pump test data into storage")

func main() {
	flag.Parse()
	// runKube()
	runRBAC()
}

func runKube() {
	ctx := context.Background()
	store := getPersistentKube(ctx, 10000, 1)

	pq, err := rego.New(
		rego.Query("data.kubernetes.validating.ingress.deny = _"),
		rego.Compiler(ast.MustCompileModules(map[string]string{"test.rego": exampleKube})),
		rego.Store(store)).PrepareForEval(ctx)
	check(err)

	go func() {
		t := time.NewTicker(time.Second)
		for range t.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			// For info on each, see: https://golang.org/pkg/runtime/#MemStats
			fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
			fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
			fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
			fmt.Printf("\tNumGC = %v\n", m.NumGC)
		}
	}()

	for {
		runKubeQuery(ctx, 10000, pq)
	}
}

func runKubeQuery(ctx context.Context, n int, pq rego.PreparedEvalQuery) {
	for i := 0; i < n; i++ {
		m := metrics.New()
		rs, err := pq.Eval(ctx, rego.EvalMetrics(m), rego.EvalInput(exampleK8sInput))
		check(err)
		if len(rs) != 1 {
			panic("undefined result")
		}
	}
}

func getPersistentKube(ctx context.Context, numNamespaces, numIngresses int) storage.Store {

	dir := "./testdata"

	if *pump {
		os.RemoveAll(dir)
		err := os.MkdirAll(dir, 0755)
		check(err)
	}

	store := persistent.New(dir, []storage.Path{
		storage.MustParsePath("/kubernetes/ingresses"),
		storage.MustParsePath("/bundles"),
		storage.MustParsePath("/system"),
	})

	if *pump {
		txn, err := store.NewTransaction(ctx, storage.WriteParams)
		check(err)

		err = store.Write(ctx, txn, storage.AddOp, storage.MustParsePath("/kubernetes/ingresses"), map[string]interface{}{})
		check(err)

		for i := 0; i < numNamespaces; i++ {
			ns := fmt.Sprintf("ns%d", i)
			for j := 0; j < numIngresses; j++ {
				name := fmt.Sprintf("obj%d", j)
				if j == 0 {
					err = store.Write(ctx, txn, storage.AddOp, storage.MustParsePath("/kubernetes/ingresses/"+ns), map[string]interface{}{})
					check(err)
				}
				n := (i * numIngresses) + j
				if n%100 == 0 {
					fmt.Println("committing", n)
					err = store.Commit(ctx, txn)
					check(err)
					txn, err = store.NewTransaction(ctx, storage.WriteParams)
					check(err)
				}
				obj := exampleK8sIngress
				err = store.Write(ctx, txn, storage.AddOp, storage.MustParsePath("/kubernetes/ingresses/"+ns+"/"+name), obj)
				check(err)
			}
		}
		err = store.Commit(ctx, txn)
		check(err)
	}

	return store
}

func runRBAC() {
	ctx := context.Background()
	store := getPersistentRBAC(ctx, 1000*1000*10)
	// store := getInmemRBAC(ctx, 1000*1000*10)

	pq, err := rego.New(
		rego.Query("data.app.rbac.allow = true"),
		rego.Compiler(ast.MustCompileModules(map[string]string{"test.rego": exampleRBAC})),
		rego.Store(store)).PrepareForEval(ctx)
	check(err)

	go func() {
		t := time.NewTicker(time.Second)
		for range t.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			// For info on each, see: https://golang.org/pkg/runtime/#MemStats
			fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
			fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
			fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
			fmt.Printf("\tNumGC = %v\n", m.NumGC)
		}
	}()

	for {
		runRBACQuery(ctx, 1000, pq)
	}
}

func runRBACQuery(ctx context.Context, n int, pq rego.PreparedEvalQuery) {
	for i := 0; i < n; i++ {
		m := metrics.New()
		rs, err := pq.Eval(ctx,
			rego.EvalMetrics(m),
			rego.EvalInput(map[string]interface{}{
				"action": "read",
				"type":   "dog",
				"user":   "alice10000",
			}))
		check(err)
		if len(rs) != 1 {
			check(errors.New("undefined result"))
		}
	}
}

func getPersistentRBAC(ctx context.Context, numUsers int) storage.Store {

	dir := "./testdata"

	if *pump {
		os.RemoveAll(dir)
		err := os.MkdirAll(dir, 0755)
		check(err)
	}

	store := persistent.New(dir, []storage.Path{
		{"bundles"},
		{"user_roles"},
		{"role_grants"},
		{"system"},
	})

	if *pump {

		txn, err := store.NewTransaction(ctx, storage.WriteParams)
		check(err)

		roleGrants := util.MustUnmarshalJSON([]byte(`{
				"customer": [
					{
						"action": "read",
						"type": "dog"
					},
					{
						"action": "read",
						"type": "cat"
					},
					{
						"action": "adopt",
						"type": "dog"
					},
					{
						"action": "adopt",
						"type": "cat"
					}
				],
				"employee": [
					{
						"action": "read",
						"type": "dog"
					},
					{
						"action": "read",
						"type": "cat"
					},
					{
						"action": "update",
						"type": "dog"
					},
					{
						"action": "update",
						"type": "cat"
					}
				],
				"billing": [
					{
						"action": "read",
						"type": "finance"
					},
					{
						"action": "update",
						"type": "finance"
					}
				]
			}`))

		err = store.Write(ctx, txn, storage.AddOp, storage.Path{"role_grants"}, roleGrants)
		check(err)

		userRoles := []interface{}{"employee", "billing"}
		err = store.Write(ctx, txn, storage.AddOp, storage.Path{"user_roles"}, map[string]interface{}{})
		check(err)

		for i := 0; i < numUsers; i++ {
			if i%100000 == 0 {
				fmt.Println("committing", i)
				err = store.Commit(ctx, txn)
				check(err)
				txn, err = store.NewTransaction(ctx, storage.WriteParams)
				check(err)
			}

			userName := fmt.Sprintf("alice%d", i)
			err = store.Write(ctx, txn, storage.AddOp, storage.Path{"user_roles", userName}, userRoles)
			check(err)
		}

		err = store.Write(ctx, txn, storage.AddOp, storage.MustParsePath("/bundles/11111111111111111111"), map[string]interface{}{
			"revision": strings.Repeat("X", 1024),
		})

		err = store.Commit(ctx, txn)
		check(err)
	}

	return store
}

func getInmemRBAC(ctx context.Context, numUsers int) storage.Store {
	store := inmem.New()

	txn, err := store.NewTransaction(ctx, storage.WriteParams)
	check(err)

	roleGrants := util.MustUnmarshalJSON([]byte(`{
			"customer": [
				{
					"action": "read",
					"type": "dog"
				},
				{
					"action": "read",
					"type": "cat"
				},
				{
					"action": "adopt",
					"type": "dog"
				},
				{
					"action": "adopt",
					"type": "cat"
				}
			],
			"employee": [
				{
					"action": "read",
					"type": "dog"
				},
				{
					"action": "read",
					"type": "cat"
				},
				{
					"action": "update",
					"type": "dog"
				},
				{
					"action": "update",
					"type": "cat"
				}
			],
			"billing": [
				{
					"action": "read",
					"type": "finance"
				},
				{
					"action": "update",
					"type": "finance"
				}
			]
		}`))

	err = store.Write(ctx, txn, storage.AddOp, storage.Path{"role_grants"}, roleGrants)
	check(err)

	fakeRoles := []interface{}{"employee", "billing"}
	userRoles := map[string]interface{}{}

	for i := 0; i < numUsers; i++ {
		userName := fmt.Sprintf("alice%d", i)
		userRoles[userName] = fakeRoles
	}

	err = store.Write(ctx, txn, storage.AddOp, storage.Path{"user_roles"}, userRoles)
	check(err)

	err = store.Commit(ctx, txn)
	check(err)

	return store
}

func runNQueriesForNTenants(ctx context.Context, n int, pq rego.PreparedEvalQuery) {
	start := n
	for i := start; i < start+n; i++ {
		tenantID := fmt.Sprintf("t%d", i)
		m := metrics.New()
		rs, err := pq.Eval(ctx,
			rego.EvalMetrics(m),
			rego.EvalInput(map[string]interface{}{
				"tenant":    tenantID,
				"operation": "op1",
			}))
		check(err)
		if len(rs) != 1 {
			check(errors.New("tenant query: undefined result"))
		}
	}
}

func runNQueriesFor1Tenants(ctx context.Context, n int, pq rego.PreparedEvalQuery) {
	for i := 0; i < n; i++ {
		tenantID := fmt.Sprintf("t%d", n)
		m := metrics.New()
		rs, err := pq.Eval(ctx,
			rego.EvalMetrics(m),
			rego.EvalInput(map[string]interface{}{
				"tenant":    tenantID,
				"operation": "op1",
			}))
		check(err)
		if len(rs) != 1 {
			check(errors.New("tenant query: undefined result"))
		}
	}
}

func getPersistentTenant(ctx context.Context) storage.Store {

	dir := "./testdata"

	if *pump {
		os.RemoveAll(dir)
		err := os.MkdirAll(dir, 0755)
		check(err)
	}

	store := persistent.New(dir, []storage.Path{
		{"tenants"},
		{"system"},
	})

	var err error

	if *pump {
		var txn storage.Transaction
		for i := 0; i < 1000*1000*10; i++ {
			if i%100000 == 0 {
				if txn != nil {
					fmt.Println("commiting", i)
					err = store.Commit(ctx, txn)
					check(err)
				}
				txn, err = store.NewTransaction(ctx, storage.WriteParams)
				check(err)
			}
			err = store.Write(ctx, txn, storage.AddOp, storage.MustParsePath(fmt.Sprintf("/tenants/t%d", i)), map[string]interface{}{
				"operations": []interface{}{
					"op1",
				},
			})
			check(err)
		}
		if txn != nil {
			fmt.Println("commiting last batch")
			err = store.Commit(ctx, txn)
			check(err)
		}
	}

	return store
}

func getInmemTenant() storage.Store {
	tenants := map[string]interface{}{}
	for i := 0; i < 1000*1000*10; i++ {
		key := fmt.Sprintf("t%d", i)
		tenants[key] = map[string]interface{}{
			"operations": []interface{}{
				"op1",
			},
		}
	}
	return inmem.NewFromObject(map[string]interface{}{
		"tenants": tenants,
	})
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

const exampleRBACData = `{
	"user_roles": {
		"alice": [
			"admin"
		],
		"bob": [
			"employee",
			"billing"
		],
		"eve": [
			"customer"
		]
	},
	"role_grants": {
		"customer": [
			{
				"action": "read",
				"type": "dog"
			},
			{
				"action": "read",
				"type": "cat"
			},
			{
				"action": "adopt",
				"type": "dog"
			},
			{
				"action": "adopt",
				"type": "cat"
			}
		],
		"employee": [
			{
				"action": "read",
				"type": "dog"
			},
			{
				"action": "read",
				"type": "cat"
			},
			{
				"action": "update",
				"type": "dog"
			},
			{
				"action": "update",
				"type": "cat"
			}
		],
		"billing": [
			{
				"action": "read",
				"type": "finance"
			},
			{
				"action": "update",
				"type": "finance"
			}
		]
	}
}
`

const exampleRBAC = `# Role-based Access Control (RBAC)
# --------------------------------
#
# This example defines an RBAC model for a Pet Store API. The Pet Store API allows
# users to look at pets, adopt them, update their stats, and so on. The policy
# controls which users can perform actions on which resources. The policy implements
# a classic Role-based Access Control model where users are assigned to roles and
# roles are granted the ability to perform some action(s) on some type of resource.
#
# This example shows how to:
#
#	* Define an RBAC model in Rego that interprets role mappings represented in JSON.
#	* Iterate/search across JSON data structures (e.g., role mappings)
#
# For more information see:
#
#	* Rego comparison to other systems: https://www.openpolicyagent.org/docs/latest/comparison-to-other-systems/
#	* Rego Iteration: https://www.openpolicyagent.org/docs/latest/#iteration

package app.rbac

# By default, deny requests.
default allow = false

# Allow admins to do anything.
allow {
	user_is_admin
}

# Allow the action if the user is granted permission to perform the action.
allow {
	# Find grants for the user.
	some grant
	user_is_granted[grant]

	# Check if the grant permits the action.
	input.action == grant.action
	input.type == grant.type
}

# user_is_admin is true if...
user_is_admin {

	some i

	data.user_roles[input.user][i] == "admin"
}

user_is_granted[grant] {
	some i, j

	role := data.user_roles[input.user][i]

	grant := data.role_grants[role][j]
}`

const exampleKube = `# Ingress Conflicts
# -----------------
#
# This example prevents conflicting Kubernetes Ingresses from being created. Two
# Kubernetes Ingress resources are considered in conflict if they have the same
# hostname. This example shows how to:
#
#	* Iterate/search across JSON arrays and objects.
#	* Leverage external context in decision-making.
#	* Define helper rules that provide useful abstractions.
#
# For additional information see:
#
#	* Rego Iteration: https://www.openpolicyagent.org/docs/latest/#iteration
# of the rules in the current package. You can evaluate specific rules by selecting
# the rule name (e.g., "deny") and clicking Evaluate Selection.

package kubernetes.validating.ingress

deny[msg] {
	# This rule only applies to Kubernetes Ingress resources.
	is_ingress
	input_host := input.request.object.spec.rules[_].host

	some other_ns, other_name
	other_host := data.kubernetes.ingresses[other_ns][other_name].spec.rules[_].host

	# Check if this Kubernetes Ingress resource is the same as the other one that
	# exists in the cluster. This is important because this policy will be applied
	# to CREATE and UPDATE operations. Resources do not conflict with themselves.
	#
	[input_ns, input_name] != [other_ns, other_name]

	# Check if there is a conflict. This check could be more sophisticated if needed.
	input_host == other_host

	# Construct an error message to return to the user.
	msg := sprintf("Ingress host conflicts with ingress %v/%v", [other_ns, other_name])
}

input_ns = input.request.object.metadata.namespace

input_name = input.request.object.metadata.name

is_ingress {
	input.request.kind.kind == "Ingress"
	input.request.kind.group == "extensions"
	input.request.kind.version == "v1beta1"
}
`

var exampleK8sInput = util.MustUnmarshalJSON([]byte(`{
    "apiVersion": "admission.k8s.io/v1beta1",
    "kind": "AdmissionReview",
    "request": {
        "kind": {
            "group": "extensions",
            "kind": "Ingress",
            "version": "v1beta1"
        },
        "operation": "CREATE",
        "userInfo": {
            "groups": null,
            "username": "alice"
        },
        "object": {
            "metadata": {
                "name": "prod",
                "namespace": "ecommerce"
            },
            "spec": {
                "rules": [
                    {
                        "host": "initech.com",
                        "http": {
                            "paths": [
                                {
                                    "path": "/finance",
                                    "backend": {
                                        "serviceName": "banking",
                                        "servicePort": 443
                                    }
                                }
                            ]
                        }
                    }
                ]
            }
        }
    }
}
`))

var exampleK8sIngress = util.MustUnmarshalJSON([]byte(`{
	"kind": "Ingress",
	"metadata": {
		"name": "foo",
		"namespace": "ecommerce"
	},
	"spec": {
		"rules": [
			{
				"host": "initech.com",
				"http": {
					"paths": [
						{
							"path": "/finance",
							"backend": {
								"serviceName": "banking",
								"servicePort": 443
							}
						}
					]
				}
			}
		]
	}
}`))
