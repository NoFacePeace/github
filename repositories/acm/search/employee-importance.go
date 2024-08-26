package search

type Employee struct {
	Id           int
	Importance   int
	Subordinates []int
}

func getImportance(employees []*Employee, id int) int {
	scores := map[int]int{}
	subs := map[int][]int{}
	for _, employee := range employees {
		id, score, sub := employee.Id, employee.Importance, employee.Subordinates
		scores[id] = score
		subs[id] = sub
	}
	sum := 0
	visit := map[int]bool{}
	queue := []int{id}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if visit[id] {
			continue
		}
		visit[id] = true
		sum += scores[id]
		for _, sub := range subs[id] {
			if visit[sub] {
				continue
			}
			queue = append(queue, sub)
		}
	}
	return sum
}
