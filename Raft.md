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

2. 获得响应后，候选人要检查自己的**state** 和 **term** 是否因为发送**RPC**而改变 (**复制日志中同样需要考虑**)

   ```go
   if rf.state != "candidate" || rf.currentTerm!= args.Term { return }
   ```

3. 若候选人获得的投票超过**半数**，则变成领导人

4. **请求投票PRC** ⭐（接收者指接收请求投票PRC的peer）

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



### 日志复制

1. 