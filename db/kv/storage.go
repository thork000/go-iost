package kv

import (
	"github.com/iost-official/go-iost/db/kv/leveldb"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/metrics"
)

var (
	dbMetrics = metrics.NewCounter("iost_db_qps", []string{"path", "func"})
)

// StorageType is the type of storage, include leveldb and rocksdb
type StorageType uint8

// Storage type constant
const (
	_ StorageType = iota
	LevelDBStorage
)

// StorageBackend is the storage backend interface
type StorageBackend interface {
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Has(key []byte) (bool, error)
	Delete(key []byte) error
	Keys(prefix []byte) ([][]byte, error)
	BeginBatch() error
	CommitBatch() error
	Size() (int64, error)
	Close() error
	NewIteratorByPrefix(prefix []byte) interface{}
}

// Storage is a kv database
type Storage struct {
	StorageBackend
	path string
}

func (s *Storage) Get(key []byte) ([]byte, error) {
	dbMetrics.Add(1, map[string]string{"path": s.path, "func": "get"})
	ilog.Debugln("dbget key:", string(key))
	return s.StorageBackend.Get(key)
}

func (s *Storage) Put(key []byte, value []byte) error {
	dbMetrics.Add(1, map[string]string{"path": s.path, "func": "put"})
	return s.StorageBackend.Put(key, value)
}

func (s *Storage) Has(key []byte) (bool, error) {
	dbMetrics.Add(1, map[string]string{"path": s.path, "func": "has"})
	return s.StorageBackend.Has(key)
}

func (s *Storage) Delete(key []byte) error {
	dbMetrics.Add(1, map[string]string{"path": s.path, "func": "delete"})
	return s.StorageBackend.Delete(key)
}

func (s *Storage) Keys(prefix []byte) ([][]byte, error) {
	dbMetrics.Add(1, map[string]string{"path": s.path, "func": "keys"})
	return s.StorageBackend.Keys(prefix)
}

func (s *Storage) BeginBatch() error {
	dbMetrics.Add(1, map[string]string{"path": s.path, "func": "begin_batch"})
	return s.StorageBackend.BeginBatch()
}

func (s *Storage) CommitBatch() error {
	dbMetrics.Add(1, map[string]string{"path": s.path, "func": "commit_batch"})
	return s.StorageBackend.CommitBatch()
}

func (s *Storage) Size() (int64, error) {
	dbMetrics.Add(1, map[string]string{"path": s.path, "func": "size"})
	return s.StorageBackend.Size()
}

func (s *Storage) Close() error {
	dbMetrics.Add(1, map[string]string{"path": s.path, "func": "close"})
	return s.StorageBackend.Close()
}

// NewStorage return the storage of the specify type
func NewStorage(path string, t StorageType) (*Storage, error) {
	switch t {
	case LevelDBStorage:
		sb, err := leveldb.NewDB(path)
		if err != nil {
			return nil, err
		}
		return &Storage{StorageBackend: sb, path: path}, nil
	default:
		sb, err := leveldb.NewDB(path)
		if err != nil {
			return nil, err
		}
		return &Storage{StorageBackend: sb, path: path}, nil
	}
}

// NewIteratorByPrefix returns a new iterator by prefix
func (s *Storage) NewIteratorByPrefix(prefix []byte) *Iterator {
	dbMetrics.Add(1, map[string]string{"path": s.path, "func": "iter"})
	ib := s.StorageBackend.NewIteratorByPrefix(prefix).(IteratorBackend)
	return &Iterator{
		IteratorBackend: ib,
	}
}

// IteratorBackend is the storage iterator backend
type IteratorBackend interface {
	Next() bool
	Key() []byte
	Value() []byte
	Error() error
	Release()
}

// Iterator is the storage iterator
type Iterator struct {
	IteratorBackend
}
