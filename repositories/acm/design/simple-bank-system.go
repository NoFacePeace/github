package design

type Bank struct {
	db map[int]int64
}

func NewBank(balance []int64) Bank {
	db := map[int]int64{}
	for k, v := range balance {
		db[k+1] = v
	}
	return Bank{
		db: db,
	}
}

func (this *Bank) Transfer(account1 int, account2 int, money int64) bool {
	if !this.exist(account1) {
		return false
	}
	if !this.exist(account2) {
		return false
	}
	if money > this.db[account1] {
		return false
	}
	this.db[account1] -= money
	this.db[account2] += money
	return true
}

func (this *Bank) Deposit(account int, money int64) bool {
	if !this.exist(account) {
		return false
	}
	this.db[account] += money
	return true
}

func (this *Bank) Withdraw(account int, money int64) bool {
	if !this.exist(account) {
		return false
	}
	if money > this.db[account] {
		return false
	}
	this.db[account] -= money
	return true
}

func (this *Bank) exist(account int) bool {
	_, ok := this.db[account]
	return ok
}

/**
 * Your Bank object will be instantiated and called as such:
 * obj := Constructor(balance);
 * param_1 := obj.Transfer(account1,account2,money);
 * param_2 := obj.Deposit(account,money);
 * param_3 := obj.Withdraw(account,money);
 */
