package design

import (
	"testing"
)

func TestRobot(t *testing.T) {
	robot := NewRobot(6, 3)
	robot.Step(2)
	robot.Step(2)
	robot.Step(2)
	robot.Step(1)
	robot.Step(4)
}
