package cache

import (
	"math/rand"
)

const (
	maxLevel = 32
)

// Level aspect of a node
type Level struct {
	forward *node // forward node has greater score
	span    int64
}

type node struct {
	Member   string
	Score    int64
	backward *node
	level    []*Level // level[0] is base level
}

type skiplist struct {
	header *node
	tail   *node
	length int32
	level  int16
}

func makeNode(level int16, score int64, member string) *node {
	n := &node{
		Score:  score,
		Member: member,
		level:  make([]*Level, level),
	}
	for i := range n.level {
		n.level[i] = new(Level)
	}
	return n
}

func makeSkiplist() *skiplist {
	return &skiplist{
		level:  1,
		header: makeNode(maxLevel, 0, ""),
	}
}

func randomLevel() int16 {
	level := int16(1)
	for float32(rand.Int31()&0xFFFF) < (0.25 * 0xFFFF) {
		level++
	}
	if level < maxLevel {
		return level
	}
	return maxLevel
}

func (skiplist *skiplist) insert(member string, score int64) {
	update := make([]*node, maxLevel) // link new node with node in `update`
	rank := make([]int64, maxLevel)

	// find position to insert
	node := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		if i == skiplist.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1] // store rank that is crossed to reach the insert position
		}
		if node.level[i] != nil {
			// traverse the skip list
			for node.level[i].forward != nil &&
				(node.level[i].forward.Score < score ||
					(node.level[i].forward.Score == score && node.level[i].forward.Member < member)) { // same score, different key
				rank[i] += node.level[i].span
				node = node.level[i].forward
			}
		}
		update[i] = node
	}

	level := randomLevel()
	// extend skiplist level
	if level > skiplist.level {
		for i := skiplist.level; i < level; i++ {
			rank[i] = 0
			update[i] = skiplist.header
			update[i].level[i].span = int64(skiplist.length)
		}
		skiplist.level = level
	}

	// make node and link into skiplist
	node = makeNode(level, score, member)
	for i := int16(0); i < level; i++ {
		node.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = node

		// update span covered by update[i] as node is inserted here
		node.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	// increment span for untouched levels
	for i := level; i < skiplist.level; i++ {
		update[i].level[i].span++
	}

	// set backward node
	if update[0] == skiplist.header {
		node.backward = nil
	} else {
		node.backward = update[0]
	}
	if node.level[0].forward != nil {
		node.level[0].forward.backward = node
	} else {
		skiplist.tail = node
	}
	skiplist.length++
}

/*
 * param node: node to delete
 * param update: backward node (of target)
 */
func (skiplist *skiplist) removeNode(node *node, update []*node) {
	for i := int16(0); i < skiplist.level; i++ {
		if update[i].level[i].forward == node {
			update[i].level[i].span += node.level[i].span - 1
			update[i].level[i].forward = node.level[i].forward
		} else {
			update[i].level[i].span--
		}
	}
	if node.level[0].forward != nil {
		node.level[0].forward.backward = node.backward
	} else {
		skiplist.tail = node.backward
	}
	for skiplist.level > 1 && skiplist.header.level[skiplist.level-1].forward == nil {
		skiplist.level--
	}
	skiplist.length--
}

/*
 * return: has found and removed node
 */
func (skiplist *skiplist) remove(member string, score int64) {
	/*
	 * find backward node (of target) or last node of each level
	 * their forward need to be updated
	 */
	update := make([]*node, maxLevel)
	node := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for node.level[i].forward != nil &&
			(node.level[i].forward.Score < score ||
				(node.level[i].forward.Score == score &&
					node.level[i].forward.Member < member)) {
			node = node.level[i].forward
		}
		update[i] = node
	}
	node = node.level[0].forward
	if node != nil && score == node.Score && node.Member == member {
		skiplist.removeNode(node, update)
	}
}

/*
 * return: 1 based rank, 0 means member not found
 */
func (skiplist *skiplist) getRank(member string, score int64) int64 {
	var rank int64 = 0
	x := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.Score < score ||
				(x.level[i].forward.Score == score &&
					x.level[i].forward.Member <= member)) {
			rank += x.level[i].span
			x = x.level[i].forward
		}

		/* x might be equal to zsl->header, so test if obj is non-NULL */
		if x.Member == member {
			return rank
		}
	}
	return 0
}

/*
 * 1-based rank
 */
func (skiplist *skiplist) getByRank(rank int64) *node {
	var i int64 = 0
	n := skiplist.header
	// scan from top level
	for level := skiplist.level - 1; level >= 0; level-- {
		for n.level[level].forward != nil && (i+n.level[level].span) <= rank {
			i += n.level[level].span
			n = n.level[level].forward
		}
		if i == rank {
			return n
		}
	}
	return nil
}

func (skiplist *skiplist) hasInRange(min int64, max int64) bool {
	// min & max = empty
	if min > max {
		return false
	}
	// min > tail
	n := skiplist.tail
	if n == nil || min > n.Score {
		return false
	}
	// max < head
	n = skiplist.header.level[0].forward
	if n == nil || max < n.Score {
		return false
	}
	return true
}

func (skiplist *skiplist) getFirstInScoreRange(min int64, max int64) *node {
	if !skiplist.hasInRange(min, max) {
		return nil
	}
	n := skiplist.header
	// scan from top level
	for level := skiplist.level - 1; level >= 0; level-- {
		// if forward is not in range than move forward
		for n.level[level].forward != nil && min > (n.level[level].forward.Score) {
			n = n.level[level].forward
		}
	}
	/* This is an inner range, so the next node cannot be NULL. */
	n = n.level[0].forward
	if max < n.Score {
		return nil
	}
	return n
}

func (skiplist *skiplist) getLastInScoreRange(min int64, max int64) *node {
	if !skiplist.hasInRange(min, max) {
		return nil
	}
	n := skiplist.header
	// scan from top level
	for level := skiplist.level - 1; level >= 0; level-- {
		for n.level[level].forward != nil && max > (n.level[level].forward.Score) {
			n = n.level[level].forward
		}
	}
	if min > (n.Score) {
		return nil
	}
	return n
}

/*
 * return members elements which score is in range [min, max). INCLUDE min, EXCLUDE max.
 * comment: NOT range (min, max) or (min, max]
 */
func (skiplist *skiplist) GetRangeByScore(min int64, max int64) (members []string) {
	members = make([]string, 0)
	// find backward nodes (of target range) or last node of each level
	node := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for node.level[i].forward != nil {
			if min <= (node.level[i].forward.Score) { // already in range
				break
			}
			node = node.level[i].forward
		}
	}

	// node is the first one within range
	node = node.level[0].forward

	// collect member of nodes in range
	for node != nil {
		if max <= (node.Score) { // already out of range
			break
		}
		member := node.Member
		members = append(members, member)
		node = node.level[0].forward
	}
	return members
}

// 1-based rank, including start, exclude stop
func (skiplist *skiplist) GetRangeByRank(start int64, stop int64) (members []string) {
	var i int64 = 0 // rank of iterator
	members = make([]string, 0)

	// scan from top level
	node := skiplist.header
	for level := skiplist.level - 1; level >= 0; level-- {
		for node.level[level].forward != nil && (i+node.level[level].span) < start {
			i += node.level[level].span
			node = node.level[level].forward
		}
	}

	i++
	node = node.level[0].forward // first node in range

	// collect member of nodes in range
	for node != nil && i < stop {
		member := node.Member
		members = append(members, member)
		node = node.level[0].forward
		i++
	}
	return members
}

// /*
//  * return removed elements
//  */
// func (skiplist *skiplist) RemoveRangeByScore(min int64, max int64) (removed []string) {
// 	update := make([]*node, maxLevel)
// 	removed = make([]string, 0)
// 	// find backward nodes (of target range) or last node of each level
// 	node := skiplist.header
// 	for i := skiplist.level - 1; i >= 0; i-- {
// 		for node.level[i].forward != nil {
// 			if min <= (node.level[i].forward.Score) { // already in range
// 				break
// 			}
// 			node = node.level[i].forward
// 		}
// 		update[i] = node
// 	}

// 	// node is the first one within range
// 	node = node.level[0].forward

// 	// remove nodes in range
// 	for node != nil {
// 		if max <= (node.Score) { // already out of range
// 			break
// 		}
// 		next := node.level[0].forward
// 		removedElement := node.Member
// 		removed = append(removed, removedElement)
// 		skiplist.removeNode(node, update)
// 		node = next
// 	}
// 	return removed
// }

// // 1-based rank, including start, exclude stop
// func (skiplist *skiplist) RemoveRangeByRank(start int64, stop int64) (removed []string) {
// 	var i int64 = 0 // rank of iterator
// 	update := make([]*node, maxLevel)
// 	removed = make([]string, 0)

// 	// scan from top level
// 	node := skiplist.header
// 	for level := skiplist.level - 1; level >= 0; level-- {
// 		for node.level[level].forward != nil && (i+node.level[level].span) < start {
// 			i += node.level[level].span
// 			node = node.level[level].forward
// 		}
// 		update[level] = node
// 	}

// 	i++
// 	node = node.level[0].forward // first node in range

// 	// remove nodes in range
// 	for node != nil && i < stop {
// 		next := node.level[0].forward
// 		removedElement := node.Member
// 		removed = append(removed, removedElement)
// 		skiplist.removeNode(node, update)
// 		node = next
// 		i++
// 	}
// 	return removed
// }
