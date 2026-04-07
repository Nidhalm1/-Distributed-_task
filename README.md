Example: using HashiCorp memberlist in Go

This small example shows how to create a memberlist node, join peers, list members and queue a broadcast.

Build

1. cd to the project directory:

   cd c:\Users\Mo\Desktop\nvProjet

2. Fetch dependencies and build:

   go mod tidy
   go build -o memberlist-node

Run

Start two nodes (separate terminals):

Terminal 1:

   .\memberlist-node.exe -name node1 -bind 127.0.0.1:7946 -http 8001

Terminal 2 (join node1):

   .\memberlist-node.exe -name node2 -bind 127.0.0.1:7947 -join 127.0.0.1:7946 -http 8002

Inspect members:

   curl http://localhost:8001/members

Queue a broadcast:

   curl http://localhost:8001/broadcast

Notes / Next steps

- Add encryption (pre-shared key) for gossip if you need confidentiality.
- Implement a proper Delegate to store/merge user metadata.
- Add tests and containerized examples for multi-node simulation.


go build -o memberlist-demo.exe
.\memberlist-demo.exe 1059