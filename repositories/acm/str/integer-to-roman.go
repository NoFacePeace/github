package str

func intToRoman(num int) string {
	strs := []string{
		"M",
		"CM",
		"D",
		"CD",
		"C",
		"XC",
		"L",
		"XL",
		"X",
		"IX",
		"V",
		"IV",
		"I",
	}
	nums := []int{
		1000,
		900,
		500,
		400,
		100,
		90,
		50,
		40,
		10,
		9,
		5,
		4,
		1,
	}
	i := 0
	str := ""
	for num != 0 {
		if num >= nums[i] {
			str += strs[i]
			num -= nums[i]
		} else {
			i++
		}
	}
	return str
}
