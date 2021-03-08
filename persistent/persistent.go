package persistent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/util"
)

// TODO(tsandall): support multi-level partitioning for use cases like k8s
// TODO(tsandall): check that writes don't escape from the defined partitions
// TODO(tsandall): how to deal w/ overwrite? need to delete existing keys...
// TODO(tsandall): assert that partitions are disjoint
// TODO(tsandall): wrap badger errors before returning from exported functions

var errNotFound = &storage.Error{Code: storage.NotFoundErr}
var errInvalidNonLeafType = &storage.Error{Code: storage.InvalidPatchErr, Message: "invalid non-leaf node type"}
var errInvalidNonLeafKey = &storage.Error{Code: storage.InvalidPatchErr, Message: "invalid non-leaf node key"}
var errInvalidPatch = &storage.Error{Code: storage.InvalidPatchErr}
var errInvalidKey = &storage.Error{Code: storage.InternalErr, Message: "invalid key"}
var errUnknownPartition = &storage.Error{Code: storage.InternalErr, Message: "unknown partition"}

func New(dir string, partitions []storage.Path) storage.Store {
	db, err := badger.Open(badger.DefaultOptions(dir))
	check(err)
	return &store{db: db, partitions: partitions, next: 1}
}

type store struct {
	db         *badger.DB
	partitions []storage.Path
	mu         sync.Mutex
	next       uint64

	storage.PolicyNotSupported
	storage.TriggersNotSupported
}

type transaction struct {
	id         uint64
	underlying *badger.Txn
}

func (txn *transaction) ID() uint64 {
	return txn.id
}

func (s *store) NewTransaction(_ context.Context, params ...storage.TransactionParams) (storage.Transaction, error) {

	var write bool

	if len(params) > 0 {
		write = params[0].Write
	}

	txn := s.db.NewTransaction(write)

	s.mu.Lock()
	id := s.next
	s.next++
	s.mu.Unlock()

	return &transaction{underlying: txn, id: id}, nil
}

func (s *store) Commit(_ context.Context, txn storage.Transaction) error {
	underlying := txn.(*transaction).underlying
	return underlying.Commit()
}

func (s *store) Abort(_ context.Context, txn storage.Transaction) {
	u := txn.(*transaction).underlying
	u.Discard()
}

func errValueUnpartionable(p storage.Path) *storage.Error {
	return &storage.Error{Code: storage.InternalErr, Message: fmt.Sprintf("value cannot be partitioned: %v", p)}
}

func (s *store) Read(_ context.Context, txn storage.Transaction, path storage.Path) (result interface{}, err error) {

	// fmt.Println("read:", path)

	key, tail, scan, err := s.partitionRead(path)

	// fmt.Println("  --> key:", string(key), "tail:", tail)
	// defer func() {
	// 	fmt.Println("  --> result:", result, "err:", err)
	// }()

	if err != nil {
		return nil, err
	}

	u := txn.(*transaction).underlying

	if scan {
		return s.readScan(u, path)
	}

	item, err := u.Get(key)
	if err != nil {
		if badger.ErrKeyNotFound == err {
			return nil, errNotFound
		}
		return nil, err
	}

	var x interface{}

	err = item.Value(func(bs []byte) error {
		return util.NewJSONDecoder(bytes.NewReader(bs)).Decode(&x)
	})

	if err != nil {
		return nil, err
	}

	return ptr(x, tail)
}

func (s *store) readScan(txn *badger.Txn, path storage.Path) (interface{}, error) {

	var prefix []byte

	if len(path) == 0 {
		prefix = []byte("/")
	} else {
		prefix = []byte(path.String() + "/") // append / to exclude substring matches
	}

	it := txn.NewIterator(badger.IteratorOptions{
		Prefix: prefix,
	})

	defer it.Close()

	result := map[string]interface{}{}

	for it.Rewind(); it.Valid(); it.Next() {
		item := it.Item()
		subpath, ok := storage.ParsePath(string(item.Key()))
		if !ok {
			return nil, errInvalidKey
		}

		subpath = subpath[len(path):]

		err := item.Value(func(bs []byte) error {

			var val interface{}

			if err := json.Unmarshal(bs, &val); err != nil {
				return err
			}

			node := result

			for i := 0; i < len(subpath)-1; i++ {
				k := subpath[i]
				next, ok := node[k]
				if !ok {
					next = map[string]interface{}{}
					node[k] = next
				}

				// NOTE(tsandall): this assertion cannot fail because the hierarchy
				// is constructed here--a panic indicates a bug in this code.
				node = next.(map[string]interface{})
			}

			node[subpath[len(subpath)-1]] = val
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *store) Write(_ context.Context, txn storage.Transaction, op storage.PatchOp, path storage.Path, value interface{}) error {
	u := txn.(*transaction).underlying
	switch op {
	case storage.AddOp:
		return s.writeAdd(u, path, value)
	case storage.ReplaceOp:
		return errors.New("not implemented: write: replace")
	case storage.RemoveOp:
		return errors.New("not implemented: write: remove")
	default:
		return errInvalidPatch
	}
}

func (s *store) writeAdd(txn *badger.Txn, path storage.Path, value interface{}) error {

	ops, err := s.partitionWriteAdd(txn, path, value)
	if err != nil {
		return err
	}

	for _, op := range ops {
		if op.delete {
			return errors.New("not implemented: write: add: deletion")
		}

		bs, err := json.Marshal(op.val)
		if err != nil {
			return err
		}
		err = txn.Set(op.key, bs)
		if err != nil {
			return err
		}
	}

	return nil
}

type partitionOp struct {
	key    []byte
	delete bool
	val    interface{}
}

func (s *store) partitionWriteAdd(txn *badger.Txn, path storage.Path, value interface{}) ([]partitionOp, error) {

	for _, p := range s.partitions {
		if p.HasPrefix(path) {
			return s.partitionWriteAddMultiple(txn, path, value)
		}
	}

	for _, p := range s.partitions {
		if path.HasPrefix(p) {
			return s.partitionWriteAddOne(txn, path, value, len(p)+1)
		}
	}

	return nil, errUnknownPartition
}

func (s *store) partitionWriteAddMultiple(txn *badger.Txn, path storage.Path, value interface{}) ([]partitionOp, error) {

	var result []partitionOp

	for _, p := range s.partitions {
		if p.HasPrefix(path) {
			x, err := ptr(value, p[len(path):])
			if err != nil {
				return nil, err
			} else if x == nil {
				continue
			}
			obj, ok := x.(map[string]interface{})
			if !ok {
				return nil, errValueUnpartionable(p)
			}
			for k, v := range obj {
				result = append(result, partitionOp{
					key: []byte(p.String() + "/" + k),
					val: v,
				})
			}
		}
	}

	return result, nil

}

func (s *store) partitionWriteAddOne(txn *badger.Txn, path storage.Path, value interface{}, index int) ([]partitionOp, error) {

	// exact match - return one operation
	if len(path) == index {
		return []partitionOp{
			{
				key: []byte(path.String()),
				val: value,
			},
		}, nil
	}

	// prefix match - return one operation but perform read-modify-write
	key := []byte(path[:index].String())
	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}

	var modified interface{}

	err = item.Value(func(bs []byte) error {

		if err := util.Unmarshal(bs, &modified); err != nil {
			return err
		}

		node := modified

		for i := index; i < len(path)-1; i++ {
			obj, ok := node.(map[string]interface{})
			if !ok {
				return errInvalidNonLeafType
			}
			node, ok = obj[path[i]]
			if !ok {
				return errInvalidNonLeafKey
			}
		}

		obj, ok := node.(map[string]interface{})
		if !ok {
			return errInvalidNonLeafType
		}

		obj[path[len(path)-1]] = value

		return nil
	})

	if err != nil {
		return nil, err
	}

	return []partitionOp{
		{
			key: key,
			val: modified,
		},
	}, nil
}

func (s *store) partitionRead(path storage.Path) ([]byte, storage.Path, bool, error) {

	for _, p := range s.partitions {

		if p.HasPrefix(path) {
			return nil, nil, true, nil
		}

		if path.HasPrefix(p) {
			return []byte(path[:len(p)+1].String()), path[len(p)+1:], false, nil
		}

	}

	return nil, nil, false, errUnknownPartition
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ptr(x interface{}, path storage.Path) (interface{}, error) {

	result := x

	for _, k := range path {
		obj, ok := result.(map[string]interface{})
		if !ok {
			return nil, nil
		}
		result, ok = obj[k]
		if !ok {
			return nil, nil
		}
	}

	return result, nil
}
