# A sharded, fault-tolerant Graph Database

## Introduction
Our project is a sharded, fault-tolerant graph database built using **Raft**. You can shard your data across multiple clusters with multiple replicas, the data is persisted on disk for high throughput in
reads and writes. Replication and fault-tolerance is done using Raft.


## What is Raft
Raft is a consensus algorithm that is designed to be easy to understand. It's equivalent to Paxos in fault-tolerance and performance.
Each server in cluster can be in one of the following three **states**:

- Leader
- Follower
- Candidate

Generally, the servers are in leader or follower state. 
**Log Entries** are numbered sequentially and contain a term number. Entry is considered committed if it has been replicated to a majority of the servers.
There is unidirectional **RPC communication**, from leader to followers. The followers never ping the leader. The leader sends AppendEntries messages to the followers with logs containing state updates. When the leader sends AppendEntries with zero logs, that’s considered a Heartbeat. Leader sends all followers Heartbeats at regular intervals.
For **Voting**, Each server persists its current term and vote, so it doesn’t end up voting twice in the same term. On receiving a RequestVote RPC, the server denies its vote if its log is more up-to-date than the candidate. It would also deny a vote, if a minimum ElectionTimeout hasn’t passed since the last Heartbeat from the leader. Otherwise, it gives a vote and resets its ElectionTimeout timer.

## What is Consistent Hashing
Consistent Hashing is a distributed hashing scheme that operates independently of the number of servers or objects in a hash table.


## What is sharding
Sharding is a way of scaling horizontally. A sharded database architecture splits a large database into several smaller databases. Each smaller component is called a shard.
![](https://i.imgur.com/nyARswB.jpg)
Instead of storing all data on a single server, we distribute it across several servers. This reduces the load on a single resource and instead distributes it equally across all the servers. This allows us to serve more requests and traffic from the growing number of customers while maintaining performance.

## Architechture

Our architechture consists of two types of nodes
1. Zero Node
2. Alpha Node
![](https://i.imgur.com/jYhBs84.jpg)

The Zero node acts like the master node, its job is to map keys to a Alpha Group and also to maintain status of each of the Alpha Groups. The job of the Zero is to also balance nodes evenly among each of the Alpha Groups.
Each Alpha group is a group of nodes that are replicated using graph, so each shard is replicated on multiple nodes, for fault-tolerance. The assumption is that with a replication factor of `K=3, 5`the failure of an entire Alpha Group is close to zero.
Due to the use of Raft for replication our system is a **CP** system.

## How we handle Sharding

Since we are building a Graph Database, which are known to be performant for `JOIN` type queries. It was of importance to us to optimise the follow operation, i.e 
Get all the `<Nodes>` that have the a `<Relation>` with a specific `<Node>`
When thinking about sharding a highly connected graph, as seen in social networking platforms such as Facebook and Twitter, you can think about various strategies. One often-used approach is to randomly choose users (nodes or vertices of the graph) and assign them to shards. 
**Example: Get all the Friends of Sanchit**
 The “random sharding” model introduces randomness in graph traversal too, meaning to get the appropriate data for a single query, there might be multiple network hops from one server to another for a single traversal, resulting in latency issues.

 
So it was imperative that all the nodes with the same predicate `<Sanchit>.<Friend>` in this case, were to be in the same group. For each predicate and its corresponding subjects and objects, there’s a single key-value pair. So that this operation could be done with very low latency.

For example this is a valid state of the Database.

```
AlphaOne:
		key = <Friends, Sanchit>
		value = <Raj, Kannav, ...>
        
		key = <Friends, Kannav> .
		value = <Sanchit, Raj, ...>
		...

AlphaTwo;
		key = <lives-in, Sanchit>
		value = <Delhi>
        
		key = <lives-in, Kannav>
		value = <Ludhiana>
		...
```


## Flow of a Query
1. Map the `Key@Relation` predicate to a alpha group(consistent hashing, so partitioning/repartitioning is easy). Zero points to the alpha group that is servring all the requests to this predicate
   `Hash(Key, Relation) = GroupID`
2. Find out the leader of the subsequent groups to get the nodes which contain the data
   `g = {node1, node5, node6}`
3. In case of a Write, contact the leader of this group and do the Write
4. Read Only operations can be handled by the Replicas themselves

## Reading and Writing data
Once the zero and the alpha groups are set up, we can start sending HTTP Requests

All the requests to graph `/<key>/<relation>`
- Method `PUT`
- Description: Store a graph Node in the system
- Request body
```
    {
        "value": [string]
    }
```
- Method `GET`
- Description: Get a list of values pointed to by the location
- Response: Array containing all the values that correspond to the query
