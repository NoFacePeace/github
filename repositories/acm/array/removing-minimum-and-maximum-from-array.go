package array

func minimumDeletions(nums []int) int {
	n := len(nums)
	if n == 0 {
		return 0
	}
	mn := nums[0]
	mnp := 0
	mx := nums[0]
	mxp := 0
	for k, v := range nums {
		if v < mn {
			mn = v
			mnp = k
		}
		if v > mx {
			mx = v
			mxp = k
		}
	}
	n1 := max(mxp+1, mnp+1)
	n2 := max(n-mxp, n-mnp)
	n3 := min(mxp+1, mnp+1) + min(n-mxp, n-mnp)
	return min(n1, n2, n3)
}
