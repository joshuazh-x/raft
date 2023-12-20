package rafttest

import (
	"context"
	"testing"

	"go.etcd.io/raft/v3"
)

func RunWithRandomFaults(t *testing.T) {
	peers := []raft.Peer{{ID: 1, Context: nil}, {ID: 2, Context: nil}, {ID: 3, Context: nil}, {ID: 4, Context: nil}, {ID: 5, Context: nil}}
	nt := newRaftNetwork(1, 2, 3, 4, 5)

	nodes := make([]*node, 0)

	for i := 1; i <= 5; i++ {
		n := startNode(uint64(i), peers, nt.nodeNetwork(uint64(i)))
		nodes = append(nodes, n)
	}

	waitLeader(nodes)

	stopc := make(chan struct{})

	for i := range nodes {
		go func(i int) {
			for {
				select {
				case <-stopc:
					return
				default:
				}

				nodes[i].Propose(context.TODO(), []byte("somedata"))
			}
		}(i)
	}

	for _, n := range nodes {
		n.stop()
	}
}
