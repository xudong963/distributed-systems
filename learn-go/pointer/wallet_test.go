package pointer

import "testing"

func TestWallet(t *testing.T) {

	wallet := Wallet{}
	//在 Go 中，当调用一个函数或方法时，参数会被复制。
	wallet.Deposit(10)
	got := wallet.Balance()
	want := 10
	if got != want {
		t.Errorf("got %d want %d", got, want)
	}
}

// error:  got 0 want 10
// why ???
// 在 Go 中，当调用一个函数或方法时，参数会被复制。 这里的balance和wallet.go 中的balance是不一样的
// 引入指针。