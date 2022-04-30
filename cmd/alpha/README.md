The worker is the one actually stores the data, that has multiple groups that are sharded


Lets say there are n number of shards

each shard lives on a subset of servers


each data/shard has multiple replicas
replicas are distributed over multiple nodes
we call each replicated shard(using raft) a group 

