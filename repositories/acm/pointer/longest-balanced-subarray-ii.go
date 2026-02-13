package pointer

func longestBalanced(nums []int) int {
	even := 0
	odd := 0
	em := map[int]int{}
	om := map[int]int{}
	n := len(nums)
	for i := 0; i < n; i++ {
		num := nums[i]
		if num%2 == 0 {
			even++
			em[num]++
			continue
		}
		odd++
		om[num]++
	}
	l, r := 0, n-1
	ans := 0
	for l < r {
		if len(em) == len(om) {
			ans = r - l + 1
			break
		}
		ln := nums[l]
		rn := nums[r]
		if len(em) > len(om) {
			if ln%2 == 0 && rn%2 == 0 {
				even--
				if em[ln] > em[rn] {
					r--
					em[rn]--
				} else {
					em[ln]--
					l++
				}
				if em[ln] == 0 {
					delete(em, ln)
				}
				if em[rn] == 0 {
					delete(em, rn)
				}
				continue
			}
			if ln%2 == 0 {
				l++
				even--
				em[ln]--
				if em[ln] == 0 {
					delete(em, ln)
				}
				continue
			}
			if rn%2 == 0 {
				r--
				even--
				em[rn]--
				if em[rn] == 0 {
					delete(em, rn)
				}
				continue
			}
			odd--
			if om[ln] > om[rn] {
				l++
				om[ln]--
				if om[ln] == 0 {
					delete(om, ln)
				}
				continue
			}
			r--
			om[rn]--
			if om[rn] == 0 {
				delete(om, rn)
			}
			continue
		}
		if ln%2 != 0 && rn%2 != 0 {
			odd--
			if om[ln] > om[rn] {
				r--
				om[rn]--
			} else {
				om[ln]--
				l++
			}
			if om[ln] == 0 {
				delete(om, ln)
			}
			if om[rn] == 0 {
				delete(om, rn)
			}
			continue
		}
		if ln%2 != 0 {
			l++
			odd--
			om[ln]--
			if om[ln] == 0 {
				delete(om, ln)
			}
			continue
		}
		if rn%2 != 0 {
			r--
			odd--
			om[rn]--
			if om[rn] == 0 {
				delete(om, rn)
			}
			continue
		}
		if em[ln] > em[rn] {
			l++
			em[ln]--
			if em[ln] == 0 {
				delete(em, ln)
			}
			even--
			continue
		}
		r--
		em[rn]--
		if em[rn] == 0 {
			delete(em, rn)
		}
		even--
		continue
	}
	return ans
}
