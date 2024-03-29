package raft

import (
	"bytes"
	"labgob"
	"labrpc"
	"log"
	"math/rand"
	"os"
	"sort"
	"sync"
	"sync/atomic"
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

// define a struct to hold information about each log entry
type LogEntry struct {
	Term int
	Command interface{}      //a log contains a series of commands, which its state machine executes in order
}


type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
	SnapShot     []byte
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]

	// state a Raft server must maintain.
	state     string
	// persistent state on all servers
	currentTerm int               // latest term server has seen (initialized to 0)
	votedFor  int                 // candidateId that received vote in current term(or null if none)
	log[]     LogEntry            // log entries. Each entry contains command for state machine and term when entry was received by leader
	// volatile state on all servers
	commitIndex int               // index of highest log entry known to be committed (initialized to 0)
	lastApplied int               // index of highest log entry applied to state machine (initialized to 0)

	// volatile state on leaders(reinitialized after elected)
	// used by toLeader()
	nextIndex[] int               // for each server, index of the next log entry to send to that server(initialized to leader last log index+1)
	matchIndex[] int              // for each server, index of highest log entry known to be replicated on server (initialized to 0)

	applyCh   chan ApplyMsg
	ch        chan bool           // reset election time
	killCh    chan bool

	LastIncludedIndex int   // the last log entry's index in snapshotting
	LastIncludedTerm int    // the last lag entry's term in snapshotting
}

//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
//invoked by candidates to gather votes

type RequestVoteArgs struct {
	// Your code	 here (2A, 2B).
	Term int                           // candidate's term
	CandidateId int                    // candidate requesting vote
	LastLogIndex int                   // index of candidate's last log entry
	LastLogTerm int                    // term of candidate's last log entry
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your code here (2A).
	Term int                           // currentTerm, for candidate to update itself
	VoteGranted bool                   // true means candidate received vote
}

// third step: define the AppendEntries RCP struct
// invoked by leader to replicate log entries; also used as heartbeat
type AppendEntriesArgs struct {
	Term int                          // leader's term
	LeaderId int                      // because in raft only leader can link to client, so follower can redirect client by leader id
		PrevLogIndex int                  // index of log entry before new ones
	PrevLogTerm  int                  // term of prevLogIndex entry
	Entries      []LogEntry           // log entries to store (empty for heartbeat)
	LeaderCommit int                  // leader already committed log's index
}

type AppendEntriesReply struct {
	Term int                          // currentTerm, for leader to update itself
	Success bool                      // true if follower contained entry matching prevLogIndex and prevLogTerm
	ConflictTerm int
	ConflictIndex int
}

type InstallSnapShotArgs struct {
	Term int
	LeaderId int
	LastIncludedIndex int
	LastIncludedTerm int
	Data []byte
}

type InstallSnapShotReply struct {
	Term int
}

func reset(ch chan bool)  {
	select {
	case <- ch:
	default:
	}
	ch <- true
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isLeader bool
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term = rf.currentTerm
	isLeader = rf.state=="leader"
	return term, isLeader
}

func (rf* Raft)encode() []byte  {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.log)
	e.Encode(rf.LastIncludedIndex)
	e.Encode(rf.LastIncludedTerm)
	return w.Bytes()
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.

func (rf *Raft) persist() {
	// Your code here (2C).
	rf.persister.SaveRaftState(rf.encode())
}


// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 {
		return
	}
	// Your code here (2C).
	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)
	if d.Decode(&rf.currentTerm) != nil  ||
		d.Decode(&rf.votedFor) != nil || d.Decode(&rf.log) != nil ||
		d.Decode(&rf.LastIncludedIndex) != nil || d.Decode(&rf.LastIncludedTerm) != nil {
		log.Fatalf("readPersist error for server: %v", rf.me)
	}
	rf.commitIndex = rf.LastIncludedIndex
	rf.lastApplied = rf.LastIncludedIndex
}

func (rf* Raft) StartSnapShot(snapShot []byte, index int)  {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if index <= rf.LastIncludedIndex {
		return
	}
	newLog := make([]LogEntry, 0)
	newLog = append(newLog, rf.log[index-rf.LastIncludedIndex:]...)
	rf.log = newLog
	rf.LastIncludedIndex = index
	rf.LastIncludedTerm = rf.log[index-rf.LastIncludedIndex].Term
	rf.persister.SaveStateAndSnapshot(rf.encode(), snapShot)
}

func (rf* Raft) changeRole(role string)  {

	switch role {
	case "follower":
		rf.state = "follower"
		rf.votedFor = -1
		rf.persist()
	case "candidate":
		rf.state = "candidate"
		rf.votedFor = rf.me
		rf.currentTerm++
		rf.persist()
		reset(rf.ch)
		//invoke vote to be leader
		go rf.electForLeader()
	case "leader":
		// only candidate can be leader
		if rf.state != "candidate" {
			return
		}
		rf.state = "leader"
		// after to be leader, something need initialize
		rf.nextIndex = make([]int, len(rf.peers))
		rf.matchIndex = make([]int, len(rf.peers))
		for i:=0; i<len(rf.matchIndex); i++ {
			rf.matchIndex[i] = -1
		}
		// initialize nextIndex for each server, it should be leader's last log index plus 1
		for i:=0; i<len(rf.nextIndex); i++ {
			rf.nextIndex[i] = rf.getLastLogIndex()+1
		}
	}
}


func (rf* Raft) logLen() int {
	return len(rf.log) + rf.LastIncludedIndex
}
func (rf* Raft) getLastLogIndex() int {
	return rf.logLen()-1
}

func (rf* Raft) getLastLogTerm() int {
	index := rf.getLastLogIndex()
	if index<rf.LastIncludedIndex {
		return -1
	}
	return rf.log[index-rf.LastIncludedIndex].Term
}

// get prevLogIndex
func (rf* Raft) getPrevLogIndex(i int) int {
	return rf.nextIndex[i]-1
}

// get prevLog term
func (rf* Raft) getPrevLogTerm(i int) int {

	index := rf.getPrevLogIndex(i)
	if index < rf.LastIncludedIndex {
		return -1
	}
	return rf.log[index-rf.LastIncludedIndex].Term
}



//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// request vote
func (rf* Raft) electForLeader() {
	rf.mu.Lock()
	args := RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateId:  rf.me,
		LastLogIndex: rf.getLastLogIndex(),
		LastLogTerm:  rf.getLastLogTerm(),
	}
	rf.mu.Unlock()
	//initial votes 1, self votes
	var votes int32 = 1
	for i:=0; i<len(rf.peers); i++ {
		// meet myself
		if i==rf.me {
			continue
		}
		go func(index int) {
			reply := &RequestVoteReply{}
			response := rf.sendRequestVote(index, &args, reply)
			if response {
				rf.mu.Lock()
				defer rf.mu.Unlock()
				// if vote fails or elect leader, reset voteCh
				// reply.Term>current term  -> to be follower
				if reply.Term > rf.currentTerm {
					rf.currentTerm = reply.Term
					rf.changeRole("follower")
					return
				}
				if rf.state != "candidate" || rf.currentTerm!= args.Term { return }
				// get vote
				if reply.VoteGranted {
					// update vote using atomic
					atomic.AddInt32(&votes, 1)
					if atomic.LoadInt32(&votes) > int32(len(rf.peers)/2) {
						rf.changeRole("leader")
						rf.appendLogEntries()
						reset(rf.ch)
					}
				}
			}
		}(i)
	}
}



//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {

	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.currentTerm < args.Term {
		rf.currentTerm = args.Term
		rf.changeRole("follower")
	}
	reply.VoteGranted = false
	//after last condition, reply.Term>=args.Term
	reply.Term = rf.currentTerm

	// 1
	if rf.currentTerm > args.Term { return }
	//rf.currentTerm==args.Term
	// 2
	if (rf.votedFor==-1 || rf.votedFor==args.CandidateId) &&
		(args.LastLogTerm > rf.getLastLogTerm() ||
			((args.LastLogTerm==rf.getLastLogTerm())&& (args.LastLogIndex>=rf.getLastLogIndex()))) {
		reply.VoteGranted = true
		rf.votedFor = args.CandidateId
		rf.state = "follower"   // rf.state can be follower or candidate
		rf.persist()
		// reset election time
		reset(rf.ch)
	}


}


func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

//AppendEntries function
func (rf* Raft) appendLogEntries() {

	for i:=0; i<len(rf.peers); i++ {
		if i == rf.me {
			continue
		}
		go func(index int) {
			rf.mu.Lock()
			if rf.state != "leader" {
				rf.mu.Unlock()
				return
			}

			if rf.nextIndex[index] - rf.LastIncludedIndex < 1 {
				rf.transmitSnapShot(index)
				return
			}
			args := AppendEntriesArgs{
				Term:         rf.currentTerm,
				LeaderId:     rf.me,
				PrevLogIndex: rf.getPrevLogIndex(index),
				PrevLogTerm:  rf.getPrevLogTerm(index),
				Entries:      append(make([]LogEntry, 0), rf.log[rf.nextIndex[index]-rf.LastIncludedIndex:]...),
				LeaderCommit: rf.commitIndex,
			}
	
			rf.mu.Unlock()

			reply := &AppendEntriesReply{}
			respond := rf.sendAppendEntries(index, &args, reply)
			rf.mu.Lock()
			if !respond || rf.state != "leader" || rf.currentTerm != args.Term{
				rf.mu.Unlock()
				return
			}
			if reply.Term > rf.currentTerm  {
				rf.currentTerm = reply.Term
				rf.changeRole("follower")
				rf.mu.Unlock()
				return
			}
			if reply.Success {
				rf.matchIndex[index] = args.PrevLogIndex + len(args.Entries)
				rf.nextIndex[index] = rf.matchIndex[index] + 1
				rf.matchIndex[rf.me] = rf.logLen() - 1
				copyMatchIndex := make([]int,len(rf.matchIndex))
				copy(copyMatchIndex,rf.matchIndex)
				sort.Sort(sort.Reverse(sort.IntSlice(copyMatchIndex)))
				N := copyMatchIndex[len(copyMatchIndex)/2]
				if N > rf.commitIndex && rf.log[N-rf.LastIncludedIndex].Term == rf.currentTerm {
					rf.commitIndex = N
					rf.apply()
				}
				rf.mu.Unlock()
				return
			} else {
				rf.nextIndex[index] = reply.ConflictIndex
				if reply.ConflictTerm != -1 {
					c := 0
					for i:=rf.LastIncludedIndex; i<rf.logLen(); i++ {
						if rf.log[i-rf.LastIncludedIndex].Term == reply.ConflictTerm {
							c = i
						}
					}
					rf.nextIndex[index] = c+1
				}
				rf.mu.Unlock()
			}
		}(i)
	}
}

// AppendEntries PRC handler
// see AppendEntries RPC in figure2
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {

	rf.mu.Lock()
	defer rf.mu.Unlock()
	// initialize AppendEntriesReply struct
	if rf.currentTerm < args.Term {
		rf.currentTerm = args.Term
		rf.changeRole("follower")
	}
	reply.Success = false
	reply.Term = rf.currentTerm
	reply.ConflictIndex = -1
	reply.ConflictTerm = -1
	reset(rf.ch)

	// reply false if term < currentTerm
	if args.Term < rf.currentTerm { return }

	if args.PrevLogIndex >=rf.LastIncludedIndex && args.PrevLogIndex < rf.logLen() {

		if args.PrevLogTerm != rf.log[args.PrevLogIndex-rf.LastIncludedIndex].Term {
			reply.ConflictTerm = rf.log[args.PrevLogIndex-rf.LastIncludedIndex].Term
			//  then search its log for the first index
			//  whose entry has term equal to conflictTerm.
			for i:=rf.LastIncludedIndex; i<rf.logLen(); i++ {
				if rf.log[i-rf.LastIncludedIndex].Term==reply.ConflictTerm {
					reply.ConflictIndex = i
					break
				}
			}
			return
		}
	}else {
		reply.ConflictIndex = rf.logLen()
		return
	}

	index := args.PrevLogIndex
	for i:=0; i<len(args.Entries); i++ {
		index++
		if index >= rf.logLen() {
			rf.log = append(rf.log, args.Entries[i:]...)
			rf.persist()
			break
		}
		if rf.log[index-rf.LastIncludedIndex].Term != args.Entries[i].Term {
			rf.log = rf.log[:index-rf.LastIncludedIndex]
			rf.log = append(rf.log, args.Entries[i:]...)
			rf.persist()
			break
		}
	}
	// if leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	if rf.commitIndex < args.LeaderCommit {
		rf.commitIndex = min(args.LeaderCommit, rf.logLen()-1)
		rf.apply()
	}
	reply.Success = true
}

// if commitIndex>lastApplied, then lastApplied+1. and apply
// log[lastApplied] to state machine
func (rf *Raft) apply() {
	rf.commitIndex = max_(rf.commitIndex, rf.LastIncludedIndex)
	rf.lastApplied = max_(rf.lastApplied, rf.LastIncludedIndex)
	for rf.commitIndex > rf.lastApplied {
		rf.lastApplied++
		currLog := rf.log[rf.lastApplied-rf.LastIncludedIndex]
		applyMsg := ApplyMsg{
			CommandValid: true,
			Command: currLog.Command,
			CommandIndex: rf.lastApplied,
			SnapShot: nil,
		}
		rf.applyCh <- applyMsg
	}
}


func (rf* Raft) sendInstallSnapShot(server int, args *InstallSnapShotArgs, reply *InstallSnapShotReply) bool {
	ok := rf.peers[server].Call("Raft.InstallSnapShot", args, reply)
	return ok
}

func (rf *Raft) InstallSnapShot(args* InstallSnapShotArgs, reply* InstallSnapShotReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply.Term = rf.currentTerm
	if args.Term < rf.currentTerm {
		return
	}
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.changeRole("follower")
	}
	reply.Term = rf.currentTerm
	reset(rf.ch)
	// 5
	if args.LastIncludedIndex <= rf.LastIncludedIndex {
		return
	}
	msg := ApplyMsg{CommandValid:false, SnapShot: args.Data}
	// 6
	if args.LastIncludedIndex < rf.logLen()-1 {
		rf.log = append(make([]LogEntry,0), rf.log[args.LastIncludedIndex-rf.LastIncludedIndex:]...)
	}else { // 7
		rf.log = []LogEntry{ {args.LastIncludedTerm, nil}}
	}

	rf.LastIncludedIndex = args.LastIncludedIndex
	rf.LastIncludedTerm = args.LastIncludedTerm
	rf.persister.SaveStateAndSnapshot(rf.encode(), args.Data)
	rf.commitIndex = max_(rf.commitIndex, rf.LastIncludedIndex)
	rf.lastApplied = max_(rf.lastApplied, rf.LastIncludedIndex)
	if rf.lastApplied > rf.LastIncludedIndex {return}
	rf.applyCh <- msg
}

func (rf* Raft) transmitSnapShot(server int)  {
	args := InstallSnapShotArgs{
		Term:              rf.currentTerm,
		LeaderId:          rf.me,
		LastIncludedIndex: rf.LastIncludedIndex,
		LastIncludedTerm:  rf.LastIncludedTerm,
		Data:              rf.persister.ReadSnapshot(),
	}
	rf.mu.Unlock()
	reply := &InstallSnapShotReply{}
	respond := rf.sendInstallSnapShot(server, &args, reply)
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if !respond || rf.state != "leader" || rf.currentTerm != args.Term {
		return
	}
	if reply.Term > rf.currentTerm {
		rf.currentTerm = reply.Term
		rf.changeRole("follower")
		return
	}

	rf.matchIndex[server] = rf.LastIncludedIndex
	rf.nextIndex[server] = rf.LastIncludedIndex + 1

	rf.matchIndex[rf.me] = rf.logLen() - 1
	copyMatchIndex := make([]int,len(rf.matchIndex))
	copy(copyMatchIndex,rf.matchIndex)
	sort.Sort(sort.Reverse(sort.IntSlice(copyMatchIndex)))
	N := copyMatchIndex[len(copyMatchIndex)/2]
	if N > rf.commitIndex && rf.log[N-rf.LastIncludedIndex].Term == rf.currentTerm {
		rf.commitIndex = N
		rf.apply()
	}
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	index := -1
	term := rf.currentTerm
	isLeader := rf.state=="leader"
	// Your code here (2B).
	if isLeader {
		index = rf.logLen()
		newLogEntry := LogEntry{
			Term:    rf.currentTerm,
			Command: command,
		}
		rf.log = append(rf.log, newLogEntry)
		rf.persist()
		rf.appendLogEntries()
	}
	return index, term, isLeader
}

//
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// Your code here, if desired.
	reset(rf.killCh)
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	rf.state = "follower"
	rf.currentTerm = 0
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.log = make([]LogEntry, 1)

	rf.votedFor = -1

	rf.applyCh = applyCh
	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))
    for i:=0; i<len(rf.matchIndex); i++ {
    	rf.matchIndex[i] = -1
	}
	rf.ch = make(chan bool, 1)
	rf.ch = make(chan bool, 1)
	rf.killCh = make(chan bool, 1)
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())
	heartbeatTime := time.Duration(100) * time.Millisecond
	//modify Make() to create a background goroutine
	go func() {
		for {
			select {
			case <-rf.killCh:
				return
			default :
			}
			electionTimeout := time.Duration(rand.Intn(200) + 300) * time.Millisecond
			rf.mu.Lock()
			state := rf.state
			rf.mu.Unlock()
			switch state {
			case "follower", "candidate":
				// if receive rpc, then break select, reset election tim
				select {
				case <-rf.ch:
				case <-time.After(electionTimeout):
					//become Candidate if time out
					rf.mu.Lock()
					rf.changeRole("candidate")
					rf.mu.Unlock()
				}
			case "leader":
				time.Sleep(heartbeatTime) // tester doesn't allow the leader send heartbeat RPCs more than ten times per second
				rf.appendLogEntries()  // leader's task is to append log entry
			}
		}
	}()
	return rf
}
