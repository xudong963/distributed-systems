package pointer

//在 Go 中，如果一个符号（例如变量、类型、函数等）是以小写符号开头，那么它在 定义它的包之外 就是私有的。

type Wallet struct {
	balance int
}

func (w *Wallet) Balance() int  {
	return w.balance
}

func (w *Wallet) Deposit(amount int) {
	w.balance += amount
}