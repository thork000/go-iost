package blockcache

import (
	"errors"
	"fmt"
	"sync"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
)

// CacheStatus ...
type CacheStatus int

const (
	// DelSingleBlockTime ...
	DelSingleBlockTime int64 = 10
)

// BCNType type of BlockCacheNode
type BCNType int

// The types of BlockCacheNode
const (
	Linked BCNType = iota
	Single
	Virtual
)

// BlockCacheNode is the implementation of BlockCacheNode
type BlockCacheNode struct { //nolint:golint
	Block        *block.Block
	Parent       *BlockCacheNode
	Children     map[*BlockCacheNode]bool
	Type         BCNType
	Number       int64
	Witness      string
	ConfirmUntil int64
	WitnessList
	Extension []byte
}

func (bcn *BlockCacheNode) addChild(child *BlockCacheNode) {
	if child != nil {
		bcn.Children[child] = true
	}
}

func (bcn *BlockCacheNode) delChild(child *BlockCacheNode) {
	delete(bcn.Children, child)
}

func (bcn *BlockCacheNode) setParent(parent *BlockCacheNode) {
	if parent != nil {
		bcn.Parent = parent
		bcn.Type = Single

		parent.addChild(bcn)
	}
}

func (bcn *BlockCacheNode) updateVirtualBCN(parent *BlockCacheNode, block *block.Block) {
	bcn.Block = block
	bcn.Number = block.Head.Number
	bcn.Witness = block.Head.Witness
	bcn.setParent(parent)
}

// NewBCN return a new block cache node instance
func NewBCN(parent *BlockCacheNode, block *block.Block) *BlockCacheNode {
	bcn := BlockCacheNode{
		Block:    block,
		Parent:   nil,
		Children: make(map[*BlockCacheNode]bool),
	}
	if block != nil {
		bcn.Number = block.Head.Number
		bcn.Witness = block.Head.Witness
	} else {
		bcn.Number = -1
	}
	bcn.setParent(parent)
	return &bcn
}

// NewVirtualBCN return a new virtual block cache node instance
func NewVirtualBCN(parent *BlockCacheNode, block *block.Block) *BlockCacheNode {
	bcn := BlockCacheNode{
		Block:    nil,
		Parent:   nil,
		Children: make(map[*BlockCacheNode]bool),
	}
	if block != nil {
		bcn.Number = block.Head.Number - 1
	}
	bcn.setParent(parent)
	bcn.Type = Virtual
	return &bcn
}

// BlockCache defines BlockCache's API
type BlockCache interface {
	Add(*block.Block) bool
	AddGenesis(*block.Block)
	Link(*block.Block)
	Del(*block.Block)
	Flush(*block.Block) error
	Find([]byte) (*block.Block, error)
	GetBlockByNumber(int64) (*block.Block, error)
	GetBlockByHash([]byte) (*block.Block, error)
	LinkedRoot() *block.Block
	Head() *block.Block
	Draw() string
	Iterator(*block.Block) <-chan *block.Block
	IteratorParent(*block.Block) <-chan *block.Block
	GetType(*block.Block) BCNType
	FindNode([]byte) (*BlockCacheNode, error)
}

// BlockCacheImpl is the implementation of BlockCache
type BlockCacheImpl struct { //nolint:golint
	linkedRoot   *BlockCacheNode
	singleRoot   *BlockCacheNode
	head         *BlockCacheNode
	hash2node    *sync.Map // map[string]*BlockCacheNode
	leaf         map[*BlockCacheNode]int64
	baseVariable global.BaseVariable
	stateDB      db.MVCCDB
}

func (bc *BlockCacheImpl) hmget(hash []byte) (*BlockCacheNode, bool) {
	rtnI, ok := bc.hash2node.Load(string(hash))
	if !ok {
		return nil, false
	}
	bcn, okn := rtnI.(*BlockCacheNode)
	if !okn {
		bc.hash2node.Delete(string(hash))
		return nil, false
	}
	return bcn, true
}

func (bc *BlockCacheImpl) hmset(hash []byte, bcn *BlockCacheNode) {
	bc.hash2node.Store(string(hash), bcn)
}

func (bc *BlockCacheImpl) hmdel(hash []byte) {
	bc.hash2node.Delete(string(hash))
}

// NewBlockCache return a new BlockCache instance
func NewBlockCache(baseVariable global.BaseVariable) (*BlockCacheImpl, error) {
	bc := BlockCacheImpl{
		linkedRoot:   NewBCN(nil, nil),
		singleRoot:   NewBCN(nil, nil),
		hash2node:    new(sync.Map),
		leaf:         make(map[*BlockCacheNode]int64),
		baseVariable: baseVariable,
		stateDB:      baseVariable.StateDB().Fork(),
	}
	bc.linkedRoot.Number = -1
	lib, err := baseVariable.BlockChain().Top()
	if err == nil {
		bc.linkedRoot = NewBCN(nil, lib)
		bc.linkedRoot.Type = Linked
		bc.singleRoot.Type = Virtual
		bc.hmset(bc.linkedRoot.Block.HeadHash(), bc.linkedRoot)
		bc.leaf[bc.linkedRoot] = bc.linkedRoot.Number

		if err := bc.updatePending(bc.linkedRoot); err != nil {
			return nil, err
		}
		bc.linkedRoot.LibWitnessHandle()
		ilog.Info("Witness Block Num:", bc.LinkedRoot().Head.Number)
		for _, v := range bc.linkedRoot.Active() {
			ilog.Info("ActiveWitness:", v)
		}
		for _, v := range bc.linkedRoot.Pending() {
			ilog.Info("PendingWitness:", v)
		}
	}
	bc.head = bc.linkedRoot
	return &bc, nil
}

// Link call this when you run the block verify after Add() to ensure add single bcn to linkedRoot
func (bc *BlockCacheImpl) Link(blk *block.Block) {
	bcn, err := bc.FindNode(blk.HeadHash())
	if err != nil {
		return
	}
	if bcn == nil {
		return
	}
	fa, ok := bc.hmget(bcn.Block.Head.ParentHash)
	if !ok || fa.Type != Linked {
		return
	}
	bcn.Type = Linked
	delete(bc.leaf, bcn.Parent)
	bc.leaf[bcn] = bcn.Number
	bc.setHead(bcn)
	if bcn.Number > bc.head.Number {
		bc.head = bcn
	}
}

func (bc *BlockCacheImpl) setHead(h *BlockCacheNode) error {
	h.CopyWitness(h.Parent)
	if h.Number%common.VoteInterval == 0 {
		if err := bc.updatePending(h); err != nil {
			return err
		}
	}
	return nil
}

func (bc *BlockCacheImpl) updatePending(h *BlockCacheNode) error {

	ok := bc.stateDB.Checkout(string(h.Block.HeadHash()))
	if ok {
		if err := h.UpdatePending(bc.stateDB); err != nil {
			ilog.Error("failed to update pending, err:", err)
			return err
		}
		if err := h.UpdateInfo(bc.stateDB); err != nil {
			ilog.Error("failed to update witness info, err:", err)
			return err
		}
	} else {
		return errors.New("failed to state db")
	}
	return nil
}

func (bc *BlockCacheImpl) updateLongest() {
	_, ok := bc.hmget(bc.head.Block.HeadHash())
	if ok {
		return
	}
	cur := bc.linkedRoot.Number
	for key, val := range bc.leaf {
		if val > cur {
			cur = val
			bc.head = key
		}
	}
}

// Add is add a block
func (bc *BlockCacheImpl) Add(blk *block.Block) bool {
	// TODO blk.Head.ParentHash maybe be in blockChain
	fa, ok := bc.hmget(blk.Head.ParentHash)
	if !ok {
		fa = NewVirtualBCN(bc.singleRoot, blk)
		bc.hmset(blk.Head.ParentHash, fa)
	}
	newNode, ok := bc.hmget(blk.HeadHash())
	if !ok {
		newNode = NewBCN(fa, blk)
		bc.hmset(blk.HeadHash(), newNode)
	} else if newNode.Type == Virtual {
		bc.singleRoot.delChild(newNode)
		newNode.updateVirtualBCN(fa, blk)
	}

	return fa.Type == Linked
}

// AddGenesis is add genesis block
func (bc *BlockCacheImpl) AddGenesis(blk *block.Block) {
	bc.linkedRoot = NewBCN(nil, blk)
	bc.linkedRoot.Type = Linked

	if err := bc.updatePending(bc.linkedRoot); err == nil {
		bc.linkedRoot.LibWitnessHandle()
	}
	bc.head = bc.linkedRoot
	bc.hmset(bc.linkedRoot.Block.HeadHash(), bc.linkedRoot)
	bc.leaf[bc.linkedRoot] = bc.linkedRoot.Number
}

func (bc *BlockCacheImpl) delNode(bcn *BlockCacheNode) {
	fa := bcn.Parent
	bcn.Parent = nil
	if bcn.Block != nil {
		bc.hmdel(bcn.Block.HeadHash())
	}
	if fa != nil {
		fa.delChild(bcn)
	}
}

// Del is delete a block
func (bc *BlockCacheImpl) Del(blk *block.Block) {
	bcn, err := bc.FindNode(blk.HeadHash())
	if err != nil {
		return
	}
	bc.del(bcn)
	bc.updateLongest()
}

func (bc *BlockCacheImpl) del(bcn *BlockCacheNode) {
	if bcn == nil {
		return
	}
	if len(bcn.Children) == 0 {
		delete(bc.leaf, bcn)
	}
	if bcn.Parent != nil && len(bcn.Parent.Children) == 1 && bcn.Parent.Type == Linked {
		bc.leaf[bcn.Parent] = bcn.Parent.Number
	}
	for ch := range bcn.Children {
		bc.del(ch)
	}
	bc.delNode(bcn)
}

func (bc *BlockCacheImpl) delSingle() {
	height := bc.linkedRoot.Number
	if height%DelSingleBlockTime != 0 {
		return
	}
	for bcn := range bc.singleRoot.Children {
		if bcn.Number <= height {
			bc.del(bcn)
		}
	}
}

func (bc *BlockCacheImpl) flush(retain *BlockCacheNode) error {
	cur := retain.Parent
	if cur != bc.linkedRoot {
		bc.flush(cur)
	}
	for child := range cur.Children {
		if child == retain {
			continue
		}
		bc.del(child)
	}
	//confirm retain to db
	if retain.Block != nil {
		err := bc.baseVariable.BlockChain().Push(retain.Block)
		if err != nil {
			ilog.Errorf("Database error, BlockChain Push err:%v", err)
			return err
		}
		ilog.Info("confirm ", retain.Number)
		err = bc.baseVariable.StateDB().Flush(string(retain.Block.HeadHash()))
		if err != nil {
			ilog.Errorf("flush mvcc error: %v", err)
			return err
		}
		bc.delNode(cur)
		retain.Parent = nil
		retain.LibWitnessHandle()
		bc.linkedRoot = retain
	}
	return nil
}

// Flush is save a block
func (bc *BlockCacheImpl) Flush(blk *block.Block) error {
	bcn, err := bc.FindNode(blk.HeadHash())
	if err != nil {
		return fmt.Errorf("failed to find block: %v", err)
	}
	if err := bc.flush(bcn); err != nil {
		return fmt.Errorf("failed to flush block cache node: %v", err)
	}
	bc.delSingle()
	bc.updateLongest()
	return nil
}

func (bc *BlockCacheImpl) FindNode(hash []byte) (*BlockCacheNode, error) {
	bcn, ok := bc.hmget(hash)
	if !ok || bcn.Type == Virtual {
		return nil, errors.New("block not found")
	}
	return bcn, nil
}

// Find is find the block
func (bc *BlockCacheImpl) Find(hash []byte) (*block.Block, error) {
	bcn, err := bc.FindNode(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to find block: %v", err)
	}
	return bcn.Block, nil
}

// GetBlockByNumber get a block by number
func (bc *BlockCacheImpl) GetBlockByNumber(num int64) (*block.Block, error) {
	it := bc.head
	for it != nil {
		if it.Number == num {
			return it.Block, nil
		}
		it = it.Parent
	}
	return nil, fmt.Errorf("block not found")
}

// GetBlockByHash get a block by hash
func (bc *BlockCacheImpl) GetBlockByHash(hash []byte) (*block.Block, error) {
	blk, err := bc.Find(hash)
	if err != nil {
		return nil, err
	}
	return blk, nil
}

// LinkedRoot return the root node
func (bc *BlockCacheImpl) LinkedRoot() *block.Block {
	return bc.linkedRoot.Block
}

// Head return head of block cache
func (bc *BlockCacheImpl) Head() *block.Block {
	if bc.head != nil {
		return bc.head.Block
	} else {
		return nil
	}
}

func (bc *BlockCacheImpl) iterator(bcn *BlockCacheNode, ch chan *block.Block) {
	ch <- bcn.Block
	for child := range bcn.Children {
		bc.iterator(child, ch)
	}
}

// Iterator is the iterator generator for blk's children
func (bc *BlockCacheImpl) Iterator(blk *block.Block) <-chan *block.Block {
	bcn, _ := bc.FindNode(blk.HeadHash())
	ch := make(chan *block.Block)
	go func(bcn *BlockCacheNode, ch chan *block.Block) {
		bc.iterator(bcn, ch)
		close(ch)
	}(bcn, ch)
	return ch
}

// Iterator is the iterator generator for blk's parent
func (bc *BlockCacheImpl) IteratorParent(blk *block.Block) <-chan *block.Block {
	bcn, _ := bc.FindNode(blk.HeadHash())
	ch := make(chan *block.Block)
	go func(bcn *BlockCacheNode, ch chan *block.Block) {
		for parent := bcn; parent != nil; parent = parent.Parent {
			ch <- parent.Block
		}
		close(ch)
	}(bcn, ch)
	return ch
}

func (bc *BlockCacheImpl) GetType(blk *block.Block) BCNType {
	bcn, _ := bc.FindNode(blk.HeadHash())
	return bcn.Type
}
