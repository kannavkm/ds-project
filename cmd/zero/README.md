The server acts similar to master in gfs/ zero in dgraph in the sense there is only one of it


Allowed Things:

Things to build 
1. http server
2. rpc server to talk to the nodes

1. Map data to groups (consistent hashing, so partitioning/repartitioning is easy)
   x = g
2. Map groups to nodes which contain the data replicated
   g = {x1, x5, x6}
3. Contact the leader of this group
4. Get/set the data

