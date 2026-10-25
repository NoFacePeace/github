package str

func countDaysTogether(arriveAlice string, leaveAlice string, arriveBob string, leaveBob string) int {
	if leaveAlice < arriveBob {
		return 0
	}
	if arriveAlice > leaveBob {
		return 0
	}
	m := map[string]int{
		"01": 31,
		"02": 28,
		"03": 31,
		"04": 30,
		"05": 31,
		"06": 30,
		"07": 31,
		"08": 31,
		"09": 30,
		"10": 31,
		"11": 30,
		"12": 31,
	}
	f := func(start string, end string) int {
		val := 0
		for k, v := range m {
			if k > start[:2] && k < end[:2] {
				val += v
			}
		}
		if start[:2] == end[:2] {
			if start[4] <= end[4] {
				val += int(end[4]-start[4]) + 1
				val += int(end[3]-start[3]) * 10
			} else {
				val += int(end[4] - '0')
				val += int('9'-start[4]) + 2
				val += int(end[3]-start[3]) * 10
				val -= 10
			}
		} else {
			val += int(end[4] - '0')
			val += int(end[3]-'0') * 10
			val += m[start[:2]] - int(start[4]-'0') + 1
			val -= int(start[3]-'0') * 10
		}
		return val
	}
	if arriveAlice <= arriveBob && leaveAlice >= leaveBob {
		return f(arriveBob, leaveBob)
	}
	if arriveBob <= arriveAlice && leaveBob >= leaveAlice {
		return f(arriveAlice, leaveAlice)
	}
	if arriveAlice <= arriveBob {
		return f(arriveBob, leaveAlice)
	}
	return f(arriveAlice, leaveBob)
}
