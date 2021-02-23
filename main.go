package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/open-policy-agent/opa/metrics"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/tsandall/opa-persistent-store-exp/persistent"
)

func getPersistentStore(pump bool) storage.Store {

	dir := "./data"

	if pump {
		os.RemoveAll(dir)

		err := os.MkdirAll(dir, 0755)
		check(err)

		db, err := badger.Open(badger.DefaultOptions(dir))
		check(err)

		var txn *badger.Txn

		for i := 0; i < 1000*1000*10; i++ {
			if i%100000 == 0 {
				if txn != nil {
					fmt.Println("commiting", i)
					err = txn.Commit()
					check(err)
				}
				txn = db.NewTransaction(true)
			}

			err = txn.SetEntry(badger.NewEntry([]byte(fmt.Sprintf("/tenants/t%d", i)), []byte(`{"operations": ["op1"]}`)))
			check(err)
		}

		err = db.Close()
		check(err)
	}

	store := persistent.NewStore(dir)
	return store
}

func getInmemStore() storage.Store {

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

func main() {

	ctx := context.Background()

	store := getInmemStore()
	// store := getPersistentStore(false)

	pq, err := rego.New(
		rego.Query("data.tenants[input.tenant].operations[_] == input.operation"),
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
		// runNQueriesFor1Tenants(ctx, 1000, pq)
		runNQueriesForNTenants(ctx, 1000, pq)
	}

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
			check(errors.New("undefined result"))
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
			check(errors.New("undefined result"))
		}
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
