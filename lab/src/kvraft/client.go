package raftkv

import (
	"labrpc"
	"time"
)
import "crypto/rand"
import "math/big"


type Clerk struct {
	servers []*labrpc.ClientEnd
	// You will have to modify this struct.
	// You will probably have to modify your Clerk to remember which server turned out to be the leader for the last RPC,
	// and send the next RPC to that server first. This will avoid wasting time searching for the leader on every RPC,
	// which may help you pass some of the tests quickly enough.
	SelectedServer int
	// One simple and fairly efficient one is to give each client a unique identifier,
	// and then have them tag each request with a monotonically increasing sequence number.
	// If a client re-sends a request, it re-uses the same sequence number.
	// Your server keeps track of the latest sequence number it has seen for each client,
	// and simply ignores any operation that it has already seen.
	id int64
	seqNum int
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	// You'll have to add code here.
	ck.id = nrand()
	ck.seqNum = 0
	return ck
}

//
// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.Get", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) Get(key string) string {

	// You will have to modify this function.
	i := ck.SelectedServer
	for {
		args := GetArgs{
			Key: key,
		}
		reply := &GetReply{}
		ok := ck.servers[i].Call("KVServer.Get", &args, &reply)
		if !ok || reply.WrongLeader {
			i = (i+1) % len(ck.servers)
			time.Sleep(time.Duration(100)*time.Millisecond)
			continue
		}else {
			ck.SelectedServer = i
			return reply.Value
		}
	}
}

//
// shared by Put and Append.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.PutAppend", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) PutAppend(key string, value string, op string) {
	// You will have to modify this function.
	args := PutAppendArgs{
		Key:   key,
		Value: value,
		Op:    op,
		Id: ck.id,
		SeqNum: ck.seqNum,
	}
	ck.seqNum ++
	i := ck.SelectedServer
	for {
		reply := &PutAppendReply{}
		ok := ck.servers[i].Call("KVServer.PutAppend", &args, &reply)

		if !ok || reply.WrongLeader {
			i = (i+1) % len(ck.servers)
			time.Sleep(time.Duration(100)*time.Millisecond)
			continue
		}else {
			ck.SelectedServer = i
			return
		}
	}
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "Append")
}
