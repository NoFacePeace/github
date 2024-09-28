package array

import (
	"testing"
)

func TestBookMyShow(t *testing.T) {
	cmds := []string{"BookMyShow", "gather", "gather", "scatter", "scatter", "gather", "scatter", "scatter", "gather", "scatter", "scatter", "gather", "scatter", "gather", "scatter", "gather", "gather"}
	args := [][]int{
		{19, 9}, {38, 8}, {27, 3}, {36, 14}, {46, 2}, {12, 5}, {12, 12}, {43, 12}, {30, 5}, {29, 6}, {37, 18}, {6, 16}, {27, 4}, {4, 17}, {14, 7}, {11, 5}, {22, 8},
	}
	var obj BookMyShow
	for idx, cmd := range cmds {
		switch cmd {
		case "BookMyShow":
			obj = NewBookMyShow(args[idx][0], args[idx][1])
		case "gather":
			obj.Gather(args[idx][0], args[idx][1])
		case "scatter":
			obj.Scatter(args[idx][0], args[idx][1])
		}
	}
}
