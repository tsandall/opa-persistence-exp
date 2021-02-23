package persistent

import (
	"bytes"
	"context"
	"errors"
	"log"

	"github.com/dgraph-io/badger"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/util"
)

func NewStore(dir string) storage.Store {

	db, err := badger.Open(badger.DefaultOptions(dir))
	check(err)

	// TODO: support arbitrary depth
	partitions := map[string]int{
		"tenants": 2,
		"system":  1,
	}

	return &store{db: db, partitions: partitions}
}

type store struct {
	db         *badger.DB
	partitions map[string]int

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
	return &transaction{underlying: txn, id: 0}, nil
}

func (s *store) Commit(_ context.Context, txn storage.Transaction) error {
	return errors.New("not implemented: commit")
}

func (s *store) Abort(_ context.Context, txn storage.Transaction) {
	u := txn.(*transaction).underlying
	u.Discard()
}

var errUnknownPartition = &storage.Error{Code: storage.InternalErr, Message: "unknown partition"}
var errNotFound = &storage.Error{Code: storage.NotFoundErr}

func (s *store) Read(_ context.Context, txn storage.Transaction, path storage.Path) (interface{}, error) {

	// fmt.Println("read:", path)

	key, tail, err := s.partitionPath(path)
	if err != nil {
		return nil, err
	}

	// fmt.Println("  --> key:", string(key), "tail:", tail)

	u := txn.(*transaction).underlying
	item, err := u.Get(key)
	if err != nil {
		if badger.ErrKeyNotFound == err {
			return nil, errNotFound
		}
		return nil, err
	}

	val, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}

	var x interface{}
	err = util.NewJSONDecoder(bytes.NewReader(val)).Decode(&x)
	if err != nil {
		return nil, err
	}

	return ptr(x, tail)
}

func (s *store) Write(_ context.Context, txn storage.Transaction, op storage.PatchOp, path storage.Path, value interface{}) error {
	return errors.New("not implemented: write")
}

func (s *store) partitionPath(path storage.Path) ([]byte, storage.Path, error) {

	if len(path) == 0 {
		return nil, nil, errUnknownPartition
	}

	idx := s.partitions[path[0]]

	return []byte(path[:idx].String()), path[idx:], nil
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
