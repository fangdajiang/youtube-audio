package util

import (
	"fmt"
	"testing"
	"time"
)

func say(s string) {
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println(s)
	}
}

func TestHelloWorld(t *testing.T) {
	// 让另外一个线程运行
	go say("world")
	go say("cat")
	// 当前线程继续运行
	say("hello")
	fmt.Println("okok")
}

func TestGoRoutine(t *testing.T) {
	go func() {
		ticker := time.Tick(time.Second)
		//time.Sleep(time.Second)
		for {
			<-ticker
			fmt.Printf("tick at %d\n", time.Now().Second())
		}
	}()
	time.Sleep(5 * time.Second)
}
