package consistent

import (
	"fmt"
	"math/rand"

	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
)

var nodesPerMachine int = 10

type Machine struct {
	Url  string
	Port string
}

type ConsistentHasher struct {
	machineNodes map[Machine][]uint32
	ring         *redblacktree.Tree
}

func NewConsistentHasher(machines []Machine) (hasher ConsistentHasher) {
	hasher.ring = &redblacktree.Tree{Comparator: utils.UInt32Comparator}
	hasher.machineNodes = make(map[Machine][]uint32)
	for _, machine := range machines {
		for i := 0; i < nodesPerMachine; i++ {
			machineKey := rand.Uint32()
			hasher.machineNodes[machine] = append(hasher.machineNodes[machine], machineKey)
			hasher.ring.Put(machineKey, machine)
		}
	}
	return
}

func (h *ConsistentHasher) AddServer(machine Machine) error {
	if _, ok := h.machineNodes[machine]; ok {
		return fmt.Errorf("Can't add machine %v, already exists.", machine)
	}
	for i := 0; i < nodesPerMachine; i++ {
		machineKey := rand.Uint32()
		h.machineNodes[machine] = append(h.machineNodes[machine], machineKey)
		h.ring.Put(machineKey, machine)
	}
	return nil
}

func (h *ConsistentHasher) RemoveServer(machine Machine) error {
	nodes, ok := h.machineNodes[machine]
	if !ok {
		return fmt.Errorf("Can't remove machine %v, doesn't already exists.", machine)
	}
	for _, machineKey := range nodes {
		h.ring.Remove(machineKey)
	}
	delete(h.machineNodes, machine)
	return nil
}

func (h *ConsistentHasher) GetServer(key uint32) (Machine, error) {
	if h.ring.Empty() {
		return Machine{}, fmt.Errorf("Can't return a server for key %v, the ring is empty.", key)
	}
	machine, ok := h.ring.Ceiling(key)
	if ok {
		return machine.Value.(Machine), nil
	}
	return h.ring.Left().Value.(Machine), nil
}
