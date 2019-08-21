package httpx

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	config := LoadConfig()
	fmt.Printf("%+v\n", *config)
}

func TestRequest(t *testing.T) {
	start := time.Now().UnixNano()
	test1(100, 100*10000)
	end := time.Now().UnixNano()
	fmt.Printf("used(ms): %v\n", (end-start)/1000000)
}

func test1(parals int, times int) {
	wg := new(sync.WaitGroup)
	for i := 0; i < parals; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < times; i++ {
				_ = "http://" + "localhost=========" + ":" + strconv.Itoa(8080) + "/gw/mul"
			}
		}()
	}
	wg.Wait()
}

func test2(parals int, times int) {
	wg := new(sync.WaitGroup)
	for i := 0; i < parals; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < times; i++ {
				sb := bytes.NewBuffer(make([]byte,0,256))
				sb.WriteString("http://")
				sb.WriteString("localhost=========")
				sb.WriteString(":")
				sb.WriteString(strconv.Itoa(8080))
				sb.WriteString("/gw/mul")
				_ = sb.String()
			}
		}()
	}
	wg.Wait()
}
