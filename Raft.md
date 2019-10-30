## Raft--易于理解的一致性算法

### 领导选举

1. **electForLeader**函数主干，候选人针对每一个peer发送**请求投票RPC**

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

2. 获得响应后，要将reply的term于candidate的currentTerm进行比较

   ```go
   if reply.Term > rf.currentTerm {
   	rf.currentTerm = reply.Term
   	rf.changeRole("follower")
   	return
   }
   ```

3. 获得响应后，候选人要检查自己的**state** 和 **term** 是否因为发送**RPC**而改变 (**复制日志中同样需要考虑**)

   ```go
   if rf.state != "candidate" || rf.currentTerm!= args.Term { return }
   ```

4. 若候选人获得的投票超过**半数**，则变成领导人

5. **请求投票PRC** ⭐（接收者指接收 *请求投票PRC* 的peer）

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



### 日志复制

1. **appendLogEntries**函数的主干，leader针对每一个peer发送**附加日志RPC**

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

2. 同**领导选举**

3. 同**领导选举**

4. **回复成功**

5. **回复不成功**

6. **附加日志RPC** ⭐

   - 几个再次明确的地方：

     - **preLogIndex**的含义：新的日志条目(s)紧随之前的索引值，是针对每一个follower而言的==nextIndex[i]-1，每一轮重试都会改变。

     - **entries[]**的含义：准备存储的日志条目；表示心跳时为空

     - **领导人获得权力**后，初始化所有的nextIndex值为自己的最后一条日志的index+1；如果一个follower的日志跟领导人的不一样，那么在附加日志PRC时的一致性检查就会失败。领导人选举成功后跟随着可能的情况

       ![1572455070974](C:\Users\96399\AppData\Roaming\Typora\typora-user-images\1572455070974.png)

       

   - reply增加 **ConflictIndex** 和 **ConflictTerm** 用于记录日志冲突index和term

   - 如果**leader的term小于接收者的currentTerm**， 则不投票

     ```go
     if args.Term < rf.currentTerm { return }
     ```

   - 接下来就三种情况

     1. **follower的日志长度比leader的短**

     2. **follower的日志长度比leader的长，且在prevLogIndex处的term相等**

     3. **follower的日志长度比leader的长，且在prevLogIndex处的term不相等**

     
     ```go
     if args.PrevLogIndex >=rf.LastIncludedIndex && args.PrevLogIndex < rf.logLen() {
     	//不进入下面的if，就是第2种情况
     	if args.PrevLogTerm != rf.log[args.PrevLogIndex-rf.LastIncludedIndex].Term {
              //第3种情况
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
         // 第1种情况
     	reply.ConflictIndex = rf.logLen()
     	return
     }
     
     // arg.PrevLogIndex是next[i]-1
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
     
     

​        

​     