package hash

import (
	"sort"
	"strconv"
	"strings"
)

func countMentions(numberOfUsers int, events [][]string) []int {
	sort.Slice(events, func(i, j int) bool {
		str1 := events[i][1]
		str2 := events[j][1]
		t1, _ := strconv.Atoi(str1)
		t2, _ := strconv.Atoi(str2)
		if t1 == t2 {
			if events[i][2] == "HERE" {
				return false
			}
			return true
		}
		return t1 < t2
	})
	mentions := map[string]int{}
	messageEvent := "MESSAGE"
	offline := map[string]int{}
	all := 0
	offlineUsers := map[string]int{}
	for _, event := range events {
		typ := event[0]
		if typ == messageEvent {
			if event[2] == "ALL" {
				all++
				continue
			}
			if event[2] == "HERE" {
				all++
				for k, v := range offlineUsers {
					ts, _ := strconv.Atoi(event[1])
					if v+60 <= ts {
						delete(offlineUsers, k)
						continue
					}
					offline[k]++
				}
				continue
			}
			ids := strings.Split(event[2], " ")
			for _, id := range ids {
				mentions[id]++
			}
			continue
		}
		ts, _ := strconv.Atoi(event[1])
		offlineUsers["id"+event[2]] = ts
	}
	arr := make([]int, numberOfUsers)
	for i := 0; i < numberOfUsers; i++ {
		id := "id" + strconv.Itoa(i)
		arr[i] = mentions[id] + all - offline[id]
	}
	return arr
}
