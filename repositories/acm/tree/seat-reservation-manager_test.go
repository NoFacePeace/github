package tree

import (
	"testing"
)

func TestSeaManger(t *testing.T) {
	cmds := []string{"SeatManager", "reserve", "reserve", "unreserve", "reserve", "reserve", "reserve", "reserve", "unreserve"}
	args := [][]int{
		{5}, {}, {}, {2}, {}, {}, {}, {}, {5},
	}
	var obj SeatManager
	for _, cmd := range cmds {
		switch cmd {
		case "SeatManager":
			obj = NewSeaManger(args[0][0])
		case "reserve":
			ret := obj.Reserve()
			t.Log(ret)
		case "unreserve":
			obj.Unreserve(args[0][0])
		default:
			t.Error("Unknown command")
		}
	}
}
