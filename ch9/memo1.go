package memo

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

type Func func(key string) (interface{}, error)

func httpGetBody(url string) (interface{}, error) {
	fmt.Printf("fetch url %s\n", url)
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

type result struct {
	value interface{}
	err   error
}

type entry struct {
	res   result
	ready chan struct{}
}

type Memo struct {
	f     Func
	cache map[string]*entry
	mu    sync.Mutex
}

func New(f Func) *Memo {
	return &Memo{
		f:     f,
		cache: make(map[string]*entry),
	}
}

func (memo *Memo) Get(key string) (interface{}, error) {
	memo.mu.Lock()
	e, ok := memo.cache[key]

	if ok {
		memo.mu.Unlock()
		// 释放锁，但是要等entry为ready的信号才能读取
		// 多个其他后续的进程都能在这里等待 信号 📶
		<-memo.cache[key].ready
		fmt.Printf("%s from cache\n", key)
		return e.res.value, e.res.err
	}

	// not exits
	e = &entry{
		ready: make(chan struct{}),
	}

	memo.cache[key] = e // 先占位，防止其他协程取到nil

	// 占位之后立即释放锁
	memo.mu.Unlock()

	// 锁的临界区不包括
	e.res.value, e.res.err = memo.f(key)
	close(e.ready) // 广播📢 ready信号
	return e.res.value, e.res.err
}

/* 知识点
- 通过深套一层的封装，使多个协程都能读到广播的chan, 作为信号。 相比于http.Get， 设置chan信号，这是使时间耗费非常少的操作
- 使用chan的是否close状态进行广播，来同步各个协程中的变量状态
- 并发编程真的是增加了不少的心智负担啊！
*/

/* memo4的结果
go test -v -run=Memo1 -race gopl/ch9
=== RUN   TestMemo1
fetch url https://example.com/1
fetch url https://example.com
https://example.com from cache
https://example.com 1.000555167s 1256 bytes
https://example.com 1.000851958s 1256 bytes
https://example.com/1 1.001003417s 1256 bytes
https://example.com/1 from cache
https://example.com/1 1.000472083s 1256 bytes
--- PASS: TestMemo1 (1.00s)
PASS
ok      gopl/ch9        1.426s
*/
