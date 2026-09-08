# Dsched - Peer-to-Peer Request Scheduler

A **decentralized request scheduler in Go**. There is no central dispatcher: every node is both a worker and a scheduler, and any node can route an incoming request to the **least-loaded node in the cluster**.

Nodes discover each other and share their load state through a **gossip protocol**, so there is no registry, no coordinator and no static configuration. Nodes can join or die at any time and the cluster keeps routing.

---

## Why

A classic load balancer is a single point of failure and a bottleneck: every request goes through it, and it must know the state of every backend. Dsched removes it. Each node keeps an approximate, gossip-propagated view of the cluster's load and forwards work accordingly.

The trade-off is explicit: the load view is **eventually consistent**, so routing is *approximately* optimal rather than perfectly optimal - and that is enough, because the cost of a slightly stale view is far lower than the cost of a central bottleneck.

---

## Architecture

```
        request                 request                 request
           │                       │                       │
           ▼                       ▼                       ▼
      ┌─────────┐             ┌─────────┐             ┌─────────┐
      │ Node A  │◄──gossip───►│ Node B  │◄──gossip───►│ Node C  │
      │         │             │         │             │         │
      │ ┌─────┐ │             │ ┌─────┐ │             │ ┌─────┐ │
      │ │load │ │             │ │load │ │             │ │load │ │
      │ │view │ │             │ │view │ │             │ │view │ │
      │ └─────┘ │             │ └─────┘ │             │ └─────┘ │
      │ ┌─────┐ │             │ ┌─────┐ │             │ ┌─────┐ │
      │ │queue│ │             │ │queue│ │             │ │queue│ │
      │ └─────┘ │             │ └─────┘ │             │ └─────┘ │
      └────┬────┘             └─────────┘             └─────────┘
           │                       ▲
           └───── forward if B is less loaded ─────────┘
```

Each node runs three loops:

1. **Server loop** - accepts requests, decides *execute locally* or *forward*.
2. **Gossip loop** - periodically picks random peers and exchanges membership + load state.
3. **Worker loop** - drains the local queue and executes tasks, updating the local load counter.

---

## Design

### Routing

On each incoming request a node compares its own load with the best-known peer load:

```
if peerLoad(best) + hysteresis < selfLoad:
        forward to best
else:
        enqueue locally
```

- **Hysteresis** prevents ping-ponging when two nodes have almost identical load.
- **Forward hop limit** - a forwarded request carries a hop counter and is executed unconditionally at the limit, so a request can never loop through the cluster.
- **Load metric** is the queue depth plus in-flight tasks, normalized by the node's declared capacity, so heterogeneous nodes are comparable.

### Gossip

- **Membership**: each node keeps a peer table `{id, addr, incarnation, state, lastSeen}`. States are `alive → suspect → dead`, with an incarnation number so a node can refute a false suspicion about itself.
- **Bootstrap**: a new node contacts one or more seed addresses, gets a full peer list back, and is then propagated to the rest by gossip. No node needs to know the whole cluster in advance.
- **Load dissemination**: load values ride along in the same gossip round, each stamped with a version so older values never overwrite newer ones.
- **Failure detection**: missed gossip rounds move a peer to `suspect`; after a timeout it becomes `dead` and is removed from the routing candidates.
- **Fanout** is `O(log N)` peers per round, which keeps message volume flat as the cluster grows.

---

## Stack

| Concern | Choice |
|---|---|
| Language | Go 1.22 |
| Transport | TCP + JSON (pluggable codec) |
| Discovery | Custom gossip / anti-entropy protocol |
| Balancing | Least-loaded with hysteresis and hop limit |
| Testing | `go test`, in-process multi-node clusters |

No external dependencies for the core: the gossip layer, the membership table and the scheduler are implemented from scratch.

---

## Repository Layout

```
.
├── cmd/
│   └── dsched/          # Node binary
├── internal/
│   ├── gossip/          # Membership table, anti-entropy, failure detection
│   ├── scheduler/       # Routing decision, hysteresis, hop limit
│   ├── queue/           # Local task queue + workers
│   └── transport/       # Wire protocol
├── test/
│   └── cluster/         # Multi-node integration tests
├── Makefile
└── README.md
```

---

## Quick Start

**Prerequisites:** Go 1.22+.

```bash
git clone https://github.com/Nidhalm1/-Distributed-_task.git
cd -Distributed-_task
go build -o bin/dsched ./cmd/dsched
```

Start a three-node cluster on one machine:

```bash
# seed node
./bin/dsched --id=A --listen=:7001 --http=:8001 --capacity=4

# joiners
./bin/dsched --id=B --listen=:7002 --http=:8002 --capacity=4 --seeds=127.0.0.1:7001
./bin/dsched --id=C --listen=:7003 --http=:8003 --capacity=8 --seeds=127.0.0.1:7001
```

Submit work to **any** node - it will be executed wherever there is capacity:

```bash
curl -X POST http://localhost:8001/submit \
  -H 'Content-Type: application/json' \
  -d '{"payload":"job-1","cost":100}'
```

Inspect a node's view of the cluster:

```bash
curl -s http://localhost:8002/peers  | jq
curl -s http://localhost:8002/status | jq
```

```json
{
  "id": "B",
  "self_load": 0.25,
  "peers": [
    { "id": "A", "load": 0.75, "state": "alive", "last_seen_ms": 180 },
    { "id": "C", "load": 0.12, "state": "alive", "last_seen_ms": 240 }
  ],
  "forwarded": 314,
  "executed_locally": 902
}
```

---

## Configuration

| Flag | Default | Description |
|---|---|---|
| `--id` | hostname | Node identity |
| `--listen` | `:7001` | Gossip / forwarding address |
| `--http` | `:8001` | Client API address |
| `--seeds` | - | Comma-separated bootstrap addresses |
| `--capacity` | `runtime.NumCPU()` | Concurrent tasks the node can run |
| `--gossip-interval` | `500ms` | Time between gossip rounds |
| `--gossip-fanout` | `3` | Peers contacted per round |
| `--suspect-timeout` | `3s` | Silence before a peer becomes `suspect` |
| `--dead-timeout` | `10s` | Silence before a peer is removed |
| `--hysteresis` | `0.10` | Load gap required to forward |
| `--max-hops` | `2` | Forward hop limit |

---

## Testing

```bash
make test           # unit tests
make test-cluster   # multi-node integration tests
make race           # go test -race
```

The integration suite spins up in-process clusters and asserts the properties that actually matter:

- **Routing** - under a skewed load, requests submitted to a saturated node end up on the least-loaded one, and no request exceeds the hop limit.
- **Discovery** - a node started with a single seed learns every other member within a bounded number of gossip rounds.
- **Join** - a node added to a running cluster starts receiving forwarded work without any restart or config change elsewhere.
- **Failure** - a killed node is marked `dead` by all survivors, disappears from routing candidates, and no request is routed to it afterwards; the cluster keeps serving.
- **Convergence** - after load stops changing, every node's view of every other node's load converges within the expected number of rounds.
- **Anti-flapping** - two nodes with near-equal load do not forward requests back and forth.

---

## Roadmap

- [ ] Task persistence and re-execution when a node dies mid-task
- [ ] Power-of-two-choices routing instead of strict least-loaded
- [ ] gRPC transport
- [ ] Prometheus metrics endpoint
- [ ] Rack / zone awareness in routing

---

## License

MIT

## Author

**Nidhal Moussa** - [GitHub](https://github.com/Nidhalm1)
