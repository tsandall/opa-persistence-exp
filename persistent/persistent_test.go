package persistent

import (
	"context"
	"reflect"
	"testing"

	"github.com/open-policy-agent/opa/storage"

	"github.com/open-policy-agent/opa/util/test"
)

func TestScan(t *testing.T) {
	test.WithTempFS(map[string]string{}, func(dir string) {
		store := NewStore(dir, []storage.Path{{"test"}, {"ignore"}})
		ctx := context.Background()
		storage.Txn(ctx, store, storage.WriteParams, func(txn storage.Transaction) error {
			err := store.Write(ctx, txn, storage.AddOp, storage.MustParsePath("/"), map[string]interface{}{
				"test": map[string]interface{}{
					"1": "a",
					"2": "b",
					"3": "c",
				},
				"ignore": map[string]interface{}{
					"x": "y",
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			val, err := store.Read(ctx, txn, storage.MustParsePath("/"))
			if err != nil {
				t.Fatal(err)
			}
			exp := map[string]interface{}{
				"ignore": map[string]interface{}{
					"x": "y",
				},
				"test": map[string]interface{}{
					"1": "a",
					"2": "b",
					"3": "c",
				},
			}
			if !reflect.DeepEqual(exp, val) {
				t.Fatalf("expected %v but got %v", exp, val)
			}
			return nil
		})
	})
}

func TestOverride(t *testing.T) {
	test.WithTempFS(map[string]string{}, func(dir string) {
		store := NewStore(dir, []storage.Path{{"test"}})
		ctx := context.Background()
		storage.Txn(ctx, store, storage.WriteParams, func(txn storage.Transaction) error {
			err := store.Write(ctx, txn, storage.AddOp, storage.MustParsePath("/test/foo"), map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "qux",
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			err = store.Write(ctx, txn, storage.AddOp, storage.MustParsePath("/test/foo/bar/corge"), "override")
			if err != nil {
				t.Fatal(err)
			}
			val, err := store.Read(ctx, txn, storage.MustParsePath("/test/foo"))
			if err != nil {
				t.Fatal(err)
			}
			exp := map[string]interface{}{
				"bar": map[string]interface{}{
					"baz":   "qux",
					"corge": "override",
				},
			}
			if !reflect.DeepEqual(exp, val) {
				t.Fatalf("expected %v but got %v", exp, val)
			}
			return nil
		})
	})
}
