package raftkv

import (
	"bytes"
	"labgob"
	"labrpc"
	"log"
	"raft"
	"strconv"
	"sync"
	"time"
)


type Op struct {
	// Your definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
	Operator string
	Key string
	Value string
	Id int64
	SeqNum int
}

type KVServer struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg
	kvDB      map[string]string

	maxraftstate int // snapshot if log grows this big

	// Your definitions here.
	mapCh   map[int] chan Op  // for each raft log entry
	idToSeq map[int64]int
	persister *raft.Persister
	killCh  chan bool
}


func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	// Your code here.
	op := Op{
		Operator: "Get",
		Key:      args.Key,
		Value:    strconv.FormatInt(nrand(), 10),
		Id:0,
		SeqNum:0,
	}
	reply.WrongLeader = true
	index, _, isleader := kv.rf.Start(op)
	if !isleader { return }
	ch := kv.getIndexCh(index)
	newOp:= checkTime(ch)
	// check identical, then get value from kvDB and return to client
	if op.Key==newOp.Key && op.Value==newOp.Value &&
		op.Operator==newOp.Operator && op.Id==newOp.Id && op.SeqNum==newOp.SeqNum {
		reply.WrongLeader = false
		kv.mu.Lock()
		reply.Value = kv.kvDB[op.Key]
		kv.mu.Unlock()
	}
}

func (kv *KVServer) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
	op := Op{
		Operator: args.Op,
		Key:   args.Key,
		Value: args.Value,
		Id: args.Id,
		SeqNum: args.SeqNum,
	}
	reply.WrongLeader = true
	index, _, isleader := kv.rf.Start(op)
	if !isleader { return }
	ch := kv.getIndexCh(index)
	newOp:= checkTime(ch)
	if newOp.Key==op.Key && newOp.Operator==op.Operator &&
		newOp.Value==op.Value && newOp.SeqNum==op.SeqNum && newOp.Id==op.Id{
		reply.WrongLeader = false
	}
}

func (kv *KVServer) getIndexCh(index int) chan Op{
	kv.mu.Lock()
	defer kv.mu.Unlock()
	_, ok := kv.mapCh[index]
	if !ok {
		kv.mapCh[index] = make(chan Op, 1)
	}
	return kv.mapCh[index]
}

// if partition, raft's leader may not commit, so ck will block
// add timeout
func checkTime(ch chan Op)Op {
	select {
	case op:=<-ch:
		return op
	case <-time.After(time.Duration(600)*time.Millisecond):
		return Op{}
	}
}

// when server launches, read snapshot
func (kv* KVServer) readSnapShot(snapShot []byte)  {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if snapShot==nil || len(snapShot)<1 { return }
	r := bytes.NewBuffer(snapShot)
	d := labgob.NewDecoder(r)
	var kvDB map[string]string
	var idToSeq map[int64]int
	if d.Decode(&kvDB) != nil || d.Decode(&idToSeq) != nil {
		log.Fatalf("read snapShot failed")
	}else {
		kv.kvDB = kvDB
		kv.idToSeq = idToSeq
	}
}

func (kv *KVServer) startSnapShot(index int) {

	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	kv.mu.Lock()
	e.Encode(kv.kvDB)
	e.Encode(kv.idToSeq)
	kv.mu.Unlock()
	kv.rf.StartSnapShot(w.Bytes(), index)
}

//
// the tester calls Kill() when a KVServer instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (kv *KVServer) Kill() {
	kv.rf.Kill()
	kv.killCh <- true
}

func (kv *KVServer) needSnapShot() bool {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	threshold := 10
	return kv.maxraftstate > 0 && kv.maxraftstate - kv.persister.RaftStateSize() < kv.maxraftstate/threshold
}

func send(ch chan Op,op Op) {
	select{
	case  <- ch:
	default:
	}
	ch <- op
}


func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *KVServer {
	// call labgob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	labgob.Register(Op{})
	kv := new(KVServer)
	kv.me = me
	kv.maxraftstate = maxraftstate
	kv.persister = persister
	// You may need initialization code here.
	kv.kvDB = make(map[string]string)
	kv.mapCh = make(map[int]chan Op)
	kv.idToSeq = make(map[int64]int)
	kv.readSnapShot(kv.persister.ReadSnapshot())
	kv.applyCh = make(chan raft.ApplyMsg)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)
	kv.killCh = make(chan bool,1)
	go func() {
		for {
			select {
			case <- kv.killCh:
				return
			case msg := <- kv.applyCh:
				if !msg.CommandValid {
					kv.readSnapShot(msg.SnapShot)
					continue
				}
				op := msg.Command.(Op)
				kv.mu.Lock()
				sn, okk := kv.idToSeq[op.Id]
				if !okk || op.SeqNum>sn {
					if op.Operator=="Put" {
						kv.kvDB[op.Key] = op.Value
					}else if op.Operator=="Append" {
						kv.kvDB[op.Key] += op.Value
					}
					kv.idToSeq[op.Id] = op.SeqNum
				}
				kv.mu.Unlock()
				ch := kv.getIndexCh(msg.CommandIndex)
				if kv.needSnapShot() {
					go kv.startSnapShot(msg.CommandIndex)
				}
				send(ch, op)
			}
		}
	}()
	return kv
}