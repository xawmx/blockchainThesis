package vpbft

import (
	"math"
)

type node struct {
	weight   float64
	outbound float64
}

// Graph holds node and edge data.
type Graph struct {
	edges map[uint64](map[uint64]float64)
	nodes map[uint64]*node
}

// NewGraph initializes and returns a new graph.
func NewGraph() *Graph {
	return &Graph{
		edges: make(map[uint64](map[uint64]float64)),
		nodes: make(map[uint64]*node),
	}
}

// Link creates a weighted edge between a source-target node pair.
// If the edge already exists, the weight is incremented.
func (self *Graph) Link(source, target uint64, weight float64) {
	self.setLinkWeight(source, target, weight, false)
}

func (self *Graph) GetLinkWeight(source, target uint64) float64 {
	return self.edges[source][target]
}

func (self *Graph) ResetLinkWeight(source, target uint64, weight float64) {
	self.setLinkWeight(source, target, weight, true)
}

func (self *Graph) setLinkWeight(source, target uint64, weight float64, override bool) {
	if _, exist := self.nodes[source]; !exist {
		self.nodes[source] = &node{
			weight:   0,
			outbound: 0,
		}
	}

	if _, exist := self.nodes[target]; !exist {
		self.nodes[target] = &node{
			weight:   0,
			outbound: 0,
		}
	}

	if _, exist := self.edges[source]; !exist {
		self.edges[source] = map[uint64]float64{}
	}

	if override {
		self.nodes[source].outbound = self.nodes[source].outbound - self.edges[source][target] + weight
		self.edges[source][target] = weight
	} else {
		self.nodes[source].outbound += weight
		self.edges[source][target] += weight
	}
}

func (self *Graph) RankWithDefault(callback func(id uint64, rank float64)) {
	self.Rank(0.85, 0.000001, callback)
}

// Rank computes the PageRank of every node in the directed graph.
// alpha (alpha) is the damping factor, usually set to 0.85.
// epsilon (epsilon) is the convergence criteria, usually set to a tiny value.
//
// This method will run as many iterations as needed, until the graph converges.
func (self *Graph) Rank(alpha, epsilon float64, callback func(id uint64, rank float64)) {
	delta := float64(1.0)
	inverse := 1 / float64(len(self.nodes))

	// Normalize all the edge weights so that their sum amounts to 1.
	for source := range self.edges {
		if self.nodes[source].outbound > 0 {
			for target := range self.edges[source] {
				self.edges[source][target] /= self.nodes[source].outbound
			}
		}
	}

	for key := range self.nodes {
		self.nodes[key].weight = inverse
	}

	for delta > epsilon {
		leak := float64(0)
		nodes := map[uint64]float64{}

		for key, value := range self.nodes {
			nodes[key] = value.weight

			if value.outbound == 0 {
				leak += value.weight
			}

			self.nodes[key].weight = 0
		}

		leak *= alpha

		for source := range self.nodes {
			for target, weight := range self.edges[source] {
				self.nodes[target].weight += alpha * nodes[source] * weight
			}

			self.nodes[source].weight += (1-alpha)*inverse + leak*inverse
		}

		delta = 0

		for key, value := range self.nodes {
			delta += math.Abs(value.weight - nodes[key])
		}
	}

	for key, value := range self.nodes {
		callback(key, value.weight)
	}
}

// Reset clears all the current graph data.
func (self *Graph) Reset() {
	self.edges = make(map[uint64](map[uint64]float64))
	self.nodes = make(map[uint64]*node)
}
