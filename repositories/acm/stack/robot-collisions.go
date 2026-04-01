package stack

import "sort"

func survivedRobotsHealths(positions []int, healths []int, directions string) []int {
	bots := []bot{}
	n := len(positions)
	for i := 0; i < n; i++ {
		b := bot{}
		b.idx = i
		b.position = positions[i]
		b.health = healths[i]
		b.direction = directions[i]
		bots = append(bots, b)
	}
	sort.Slice(bots, func(a, b int) bool {
		return bots[a].position < bots[b].position
	})
	stack := []int{}
	idxs := []int{}
	for i := 0; i < n; i++ {
		if bots[i].direction == 'R' {
			stack = append(stack, i)
			continue
		}
		if len(stack) == 0 {
			idxs = append(idxs, i)
			continue
		}
		for len(stack) != 0 {
			idx := stack[len(stack)-1]
			if bots[idx].health > bots[i].health {
				bots[idx].health--
				break
			}
			if bots[idx].health == bots[i].health {
				stack = stack[:len(stack)-1]
				break
			}
			bots[i].health--
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				idxs = append(idxs, i)
			}
		}
	}
	for i := 0; i < len(stack); i++ {
		idxs = append(idxs, stack[i])
	}
	alive := []bot{}
	for i := 0; i < len(idxs); i++ {
		alive = append(alive, bots[idxs[i]])
	}
	sort.Slice(alive, func(a, b int) bool {
		return alive[a].idx < alive[b].idx
	})
	ans := []int{}
	for i := 0; i < len(alive); i++ {
		ans = append(ans, alive[i].health)
	}
	return ans
}

type bot struct {
	idx       int
	position  int
	health    int
	direction byte
}
