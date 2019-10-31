## Fault-tolerant Key/Value Service

**1. 结构图**

![](https://github.com/DreaMer963/distributed-systems/blob/master/pic/kvserver.png)

**2. 过程**

每一个server都关联着一个raft，client通过clerk与server交流。clerk发送put/append/get RPC给server，server将operator发送给raft，raft会返回其状态是不是leader。如果不是，重新发送RPC,直到是。然后阻塞等待，直到大部分server的raft的log提交了该entry。然后leader 将apply该entry，该server从管道读取到msg，执行操作，阻塞结束，然后将结果返回给client。