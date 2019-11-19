---
title: Raft--易于理解的一致性算法
date: 2019-11-07T00:00:00+23:00
---

这篇文章讲解如何用go实现raft算法，代码框架来自于Mit6.824分布式课程

最初，为了学习分布式系统，我了解到了 **Mit 6.824课程**，课程的 lab 需要用 go 来完成。于是 go 走进了我的世界，go 很容易入门， 写起来很舒服，但是要真正的理解 go， 并不是很容易， 特别是对 **goroutine, select, channel** 等的灵活运用。俗话说的好，初生牛犊不怕虎， 在初步了解 go 之后， 我就开始肝课程了。 每个 lab 都会有对应的论文， 比如 mapreduce, raft 等等。 lab2 是**实现 raft** 算法， lab3 是基于 lab2的 raft 算法来 实现一个简单的**分布式 kv 存储**。 在做 lab 的过程中，不仅仅可以对 raft 的细节有更好的把握, 同时对 go 语言的理解也会逐渐加深， 特别是并发部分。

首先看一下**复制状态机**,如下图所示
![复制状态机](https://github.com/DreaMer963/distributed-systems/blob/master/pic/%E5%A4%8D%E5%88%B6%E7%8A%B6%E6%80%81%E6%9C%BA.png)
复制状态机通常是基于复制日志实现的。每一台服务器存储着一个包含一系列指令的日志，每个日志都按照相同的顺序包含相同的指令，所以每一个服务器都执行相同的指令序列。那么如何保证每台服务器上的日志都相同呢？ 这就是接下来要介绍的一致性算法raft要做的事情了。

raft 主要分三大部分， **领导选举**， **日志复制**， **日志压缩**。 由于其中的细节很多，所以在实现过程中肯定会遇到各种各样的问题， 这也是一个比较好的事情，因为问题将促使我们不断地深入的去阅读论文， 同时锻炼 debug 并发程序的能力。最后肯定是满满的收获。

实现主要依赖于raft论文中的下图
![](https://github.com/DreaMer963/distributed-systems/blob/master/pic/%E5%9B%BE%E4%BA%8C.png)
代码框架条理清楚。包含七个主要的 **struct** , 三个 **RPC handler** , 四个**主干函数** , 如下
```go

// 七个struct
type Raft struct {
    ...
}

type RequestVoteArgs struct {
    ...
}

type RequestVoteReply struct {
    ...
}

type AppendEntriesArgs struct {
    ...
}

type AppendEntriesReply struct {
    ...
}

type InstallSnapShotArgs struct {
    ...
}

type InstallSnapShotReply struct {
    ...
}

// 三个 RPC handler

// RequestVote RPC handler
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
    ...
}

// AppendEntries RPC handler
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
    ...
}

// InstallSnapShot RPC handler
func (rf *Raft) InstallSnapShot(args* InstallSnapShotArgs, reply* InstallSnapShotReply) {
    ...
}

// 四个主干函数
func (rf* Raft) electForLeader() {
    ...
}

func (rf* Raft) appendLogEntries() {
    ...
}

func (rf* Raft) transmitSnapShot(server int)  {
    ...
}

// 包含一个后台 goroutine. 
func Make(peers []*labrpc.ClientEnd, me int, persister *Persister, applyCh chan ApplyMsg) *Raft {
        ...
        go func() {
            for {
                electionTimeout := time.Duration(rand.Intn(200) + 300) * time.Millisecond
                switch state {
                case "follower", "candidate":
                    // if receive rpc, then break select, reset election tim
                    select {
                    case <-rf.ch:
                    case <-time.After(electionTimeout):
                    //become Candidate if time out
                        rf.changeRole("candidate")
                    }
                case "leader":
                    time.Sleep(heartbeatTime) 
                    rf.appendLogEntries()
                }
            }
        }()
}

```

下面的部分， 记录了三大部分的主干框架和我认为的容易出错的地方。 在你实现的过程中，如果真的 debug 不出错在哪里，可以看看我下面提到的一些要点。

#### 领导选举

------

**1**. **electForLeader**函数主干，候选人针对每一个peer发送**请求投票RPC**

   ```go
   for i:=0; i<len(rf.peers); i++ {
       // meet myself
       if i==rf.me {
           continue
       }
       go func(index int) {
           reply := &RequestVoteReply{}
           response := rf.sendRequestVote(index, &args, reply)
           if response {
               ...
           }
       }(i)
   }
   ```

**2**. 获得响应后，要将reply的term于candidate的currentTerm进行比较

   ```go
   if reply.Term > rf.currentTerm {
   	rf.currentTerm = reply.Term
   	rf.changeRole("follower")
   	return
   }
   ```

**3**. 获得响应后，候选人要检查自己的 **state** 和 **term** 是否因为发送**RPC**而改变 

   ```go
   if rf.state != "candidate" || rf.currentTerm!= args.Term { return }
   ```

**4**. 若候选人获得的投票超过**半数**，则变成领导人

**5**. **请求投票PRC** ⭐（接收者指接收 *请求投票PRC* 的peer）

- 如果**candidate的term小于接收者的currentTerm**， 则不投票，并且返回接收者的currentTerm

     ```go
     reply.VoteGranted = false
     reply.Term = rf.currentTerm
     if rf.currentTerm > args.Term { return }
     ```

- 如果**接收者的votedFor为空或者为candidateId，并且candidate的日志至少和接收者一样新**，那么就投票给候选人。candidate的日志至少和接收者一样新的含义：**candidate的最后一个日志条目的term大于接收者的最后一个日志条目的term或者当二者相等时，candidate的最后一个日志条目的index要大于等于接收者的**

     ```go
     if (rf.votedFor==-1 || rf.votedFor==args.CandidateId) &&
        (args.LastLogTerm > rf.getLastLogTerm() ||
      ((args.LastLogTerm==rf.getLastLogTerm())&& (args.LastLogIndex>=rf.getLastLogIndex()))) {
            reply.VoteGranted = true
            rf.votedFor = args.CandidateId
            rf.state = "follower"   // rf.state can be follower or candidate
            ...
     }
     ```

------

#### 日志复制

**1**. **appendLogEntries** 函数的主干，leader针对每一个peer发送**附加日志RPC**

   ```go
   for i:=0; i<len(rf.peers); i++ {
       if i == rf.me {
           continue
       }
       go func(index int) {
           reply := &AppendEntriesReply{}
           respond := rf.sendAppendEntries(index, &args, reply)
           if reply.Success {
               ...
               return
           } else {
               ...
           }
       }(i)
   }
   ```

**2**. 获得响应后，要将 reply 的 term 于 leader 的 currentTerm 进行比较

**3**. 获得响应后，候选人要检查自己的 **state** 和 **term** 是否因为发送**RPC** 而改变

**4**. **回复成功**

- 更新nextIndex, matchIndex
- 如果存在一个N满足**N>commitIndex**，并且大多数**matchIndex[i] > N** 成立，并且**log[N].term == currentTerm**，则更新**commitIndex=N**

**5**. **回复不成功**

- 更新**nextIndex**，然后重试

**6**. **附加日志RPC** ⭐

- 几个再次明确的地方：

- **preLogIndex**的含义：新的日志条目(s)紧随之前的索引值，是针对每一个follower而言的==nextIndex[i]-1，每一轮重试都会改变。

- **entries[]** 的含义：准备存储的日志条目；表示心跳时为空

    ```go
    append(make([]LogEntry, 0), rf.log[rf.nextIndex[index]-rf.LastIncludedIndex:]...)
    ```

- **领导人获得权力**后，初始化所有的nextIndex值为自己的最后一条日志的index+1；如果一个follower的日志跟领导人的不一样，那么在附加日志PRC时的一致性检查就会失败。领导人选举成功后跟随着可能的情况

![](https://github.com/DreaMer963/distributed-systems/blob/master/pic/appendLog.jpg)

- reply增加 **ConflictIndex** 和 **ConflictTerm** 用于记录日志冲突index和term

- 如果**leader的term小于接收者的currentTerm**， 则不投票

- 接下来就三种情况

    1. **follower的日志长度比leader的短**
    2. **follower的日志长度比leader的长，且在prevLogIndex处的term相等**
    3. **follower的日志长度比leader的长，且在prevLogIndex处的term不相等**
    ```go
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
    ```

- 如果 **leaderCommit > commitIndex**，令 commitIndex等于leaderCommit 和新日志条目索引值中较小的一个

------


#### 日志压缩

**1**. 增量压缩的方法，这个方法每次只对一小部分数据进行操作，这样就分散了压缩的负载压力

**2**. ![](https://github.com/DreaMer963/distributed-systems/blob/master/pic/log-compact.jpg)

**3**. **安装快照RPC**

- 尽管服务器通常都是独立的创建快照，但是领导人必须偶尔的发送快照给一些落后的跟随者

- 三种情况

    - leader 的 **LastIncludedIndex** 小于等于follower的 **LastIncludeIndex**
    - leader的 **LastIncludedIndex** 大于follower的 **LastIncludeIndex**，leader 的 **LastIncludedIndex** 小于 follower 日志的最大索引值
    - leader 的 **LastIncludedIndex** 大于等于 follower 日志的最大索引值

    **对应的处理方式**
    - 如果接收到的快照是自己日志的前面部分，那么快照包含的条目将全部被删除，但是快照后面的条目仍然有效，要保留
    - 如果快照中包含没有在接收者日志中存在的信息，那么跟随者丢弃其整个日志，全部被快照取代。

