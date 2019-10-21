package raftkv

import (
	"bytes"
	"labgob"
	"labrpc"
	"log"
	"os"
	"raft"
	"strconv"
	"sync"
	"time"
)

var info *log.Logger
func init() {
	_, err := os.OpenFile("infoFile.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 6666)
	if err != nil {
		log.Fatalln("fail to open log: ", err)
	}
	info = log.New(os.Stdout, "Info: ",log.Ltime|log.Lshortfile)
}

const Debug = 0

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}


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
	}
	_, isLeader := kv.rf.GetState()
	reply.WrongLeader = true
	if !isLeader { return }
	index, _, isleader := kv.rf.Start(op)
	if !isleader { return }
	ch := kv.getIndexCh(index)
	newOp:= checkTime(ch)
	// check identical, then get value from kvDB and return to client
	if op.Key==newOp.Key && op.Value==newOp.Value && op.Operator==newOp.Operator {
		reply.WrongLeader = false
		kv.mu.Lock()
		reply.Value = kv.kvDB[op.Key]
		kv.mu.Unlock()
		return
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
	_, isLeader := kv.rf.GetState()
	if !isLeader { return }
	index, _, isleader := kv.rf.Start(op)
	if !isleader { return }
	ch := kv.getIndexCh(index)
	newOp:= checkTime(ch)
	if newOp.Key==op.Key && newOp.Operator==op.Operator &&
		newOp.Value==op.Value && newOp.SeqNum==op.SeqNum && newOp.Id==op.Id{
		reply.WrongLeader = false
		return
	}
}

func (kv *KVServer) getIndexCh(index int) chan Op{
	kv.mu.Lock()
	defer kv.mu.Unlock()
	_, ok := kv.mapCh[index]
	if !ok {
		kv.mapCh[index] = make(chan Op, 1)
	}
	ch := kv.mapCh[index]
	return ch
}

// if partition, raft's leader may not commit, so ck will block
// add timeout
func checkTime(ch chan Op)Op {
	select {
	case op:=<-ch:
		return op
	case <-time.After(time.Second):
		return Op{}
	}
}

// when server launches, read snapshot
func (kv* KVServer) readSnapShot(snapShot []byte)  {
	if len(snapShot)<1 { return }
	r := bytes.NewBuffer(snapShot)
	d := labgob.NewDecoder(r)

	if d.Decode(&kv.kvDB) != nil && d.Decode(&kv.idToSeq) != nil {
		log.Fatalf("read snapShot failed")
	}
}

func (kv *KVServer) startSnapShot(index int) {

	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	kv.mu.Lock()
	e.Encode(kv.kvDB)
	e.Encode(kv.mapCh)
	kv.mu.Unlock()
	info.Printf("进入rf")
	kv.rf.StartSnapShot(w.Bytes(), index)
}

//
// the tester calls Kill() when a KVServer instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (kv *KVServer) Kill() {
	//info.Println("开始kill....")
	kv.rf.Kill()
	//info.Println("kill kvserver")
	// Your code here, if desired.
	kv.killCh <- true
	//info.Println("kv.killCh有东西了")

}


//
// servers[] contains the ports of the set of
// servers that will cooperate via Raft to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
// the k/v server should store snapshots through the underlying Raft
// implementation, which should call persister.SaveStateAndSnapshot() to
// atomically save the Raft state along with the snapshot.
// the k/v server should snapshot when Raft's saved state exceeds maxraftstate bytes,
// in order to allow Raft to garbage-collect its log. if maxraftstate is -1,
// you don't need to snapshot.
// StartKVServer() must return quickly, so it should start goroutines
// for any long-running work.
//
func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *KVServer {
	// call labgob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	labgob.Register(Op{})

	kv := new(KVServer)
	kv.me = me
	kv.maxraftstate = maxraftstate

	// You may need initialization code here.
	kv.persister = persister
	kv.applyCh = make(chan raft.ApplyMsg)
	kv.readSnapShot(kv.persister.ReadSnapshot())
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)
	kv.mapCh = make(map[int] chan Op)
	kv.kvDB = make(map[string]string)
	kv.idToSeq = make(map[int64]int)
	kv.killCh = make(chan bool, 1)
	go func() {
		for {

			select {
			case <- kv.killCh:
				//info.Println("kill 成功")
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
					if op.Operator == "Put" {
						kv.kvDB[op.Key] = op.Value
					}else if op.Operator == "Append" {
						kv.kvDB[op.Key] += op.Value
					}
					kv.idToSeq[op.Id] = op.SeqNum
				}
				kv.mu.Unlock()
				// judge if need to start snapshot
				var threshold = int(1.5*float64(kv.maxraftstate))
				if kv.maxraftstate != -1 && kv.persister.RaftStateSize() >= threshold {
					info.Println("开始snapShot")
					go kv.startSnapShot(msg.CommandIndex)
				}

				ch := kv.getIndexCh(msg.CommandIndex)
				ch <- op
			}
		}
	}()
	return kv
}