package memo

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

func TestMemo1(t *testing.T) {
	urls := []string{
		"https://example.com",
		"https://example.com/1",
		"https://example.com",
		"https://example.com/1",
	}

	m := New(httpGetBody)
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)

		go func(url string) {
			start := time.Now()
			v, err := m.Get(url)
			if err != nil {
				log.Print(err)
			}
			fmt.Printf("%s %s %d bytes\n", url, time.Since(start), len(v.([]byte)))

			wg.Done()
		}(url)

	}

	wg.Wait()
}
