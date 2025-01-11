package array

type ATM struct {
	store [5]int
}

func Constructor() ATM {
	return ATM{
		store: [5]int{},
	}
}

func (this *ATM) Deposit(banknotesCount []int) {
	for k, v := range banknotesCount {
		this.store[k] += v
	}
}

func (this *ATM) Withdraw(amount int) []int {
	ans := make([]int, 5)
	num := min(amount/500, this.store[4])
	ans[4] = num
	amount -= num * 500

	num = min(amount/200, this.store[3])
	ans[3] = num
	amount -= num * 200

	num = min(amount/100, this.store[2])
	ans[2] = num
	amount -= num * 100

	num = min(amount/50, this.store[1])
	ans[1] = num
	amount -= num * 50

	num = min(amount/20, this.store[0])
	ans[0] = num
	amount -= num * 20

	if amount != 0 {
		return []int{-1}
	}
	for k, v := range ans {
		this.store[k] -= v
	}
	return ans
}

/**
 * Your ATM object will be instantiated and called as such:
 * obj := Constructor();
 * obj.Deposit(banknotesCount);
 * param_2 := obj.Withdraw(amount);
 */
