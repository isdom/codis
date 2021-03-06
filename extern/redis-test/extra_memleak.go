// Copyright 2014 Wandoujia Inc. All Rights Reserved.
// Licensed under the MIT (MIT-LICENSE.txt) license.

package main

import (
	"flag"
	"fmt"
	"time"
)

type ExtraMemleakTestCase struct {
	proxy string
	group int
	nkeys int
}

func init() {
	testcase = &ExtraMemleakTestCase{}
}

func (tc *ExtraMemleakTestCase) init() {
	flag.StringVar(&tc.proxy, "proxy", "", "redis host:port")
	flag.IntVar(&tc.group, "group", 8, "# of test players")
	flag.IntVar(&tc.nkeys, "nkeys", 1000, "# of keys per test")
}

func (tc *ExtraMemleakTestCase) main() {
	fmt.Println(`
!! PLEASE MAKE SURE !!
- compile : make MALLOC=libc -j4
- run     : valgrind --leak-check=full
`)
	go func() {
		c := NewConn(tc.proxy)
		for {
			time.Sleep(time.Second * 5)
			c.Check()
		}
	}()
	tg := &TestGroup{}
	tg.Reset()
	for g := 0; g < tc.group; g++ {
		tg.AddPlayer()
		go tc.player(g, tg)
	}
	tg.Start()
	tg.Wait()
	fmt.Println("done")
}

func (tc *ExtraMemleakTestCase) player(gid int, tg *TestGroup) {
	tg.PlayerWait()
	defer tg.PlayerDone()
	c := NewConn(tc.proxy)
	defer c.Close()
	r := &Rand{time.Now().UnixNano()}
	us := UnitSlice(make([]*Unit, tc.nkeys))
	for i := 0; i < len(us); i++ {
		key := fmt.Sprintf("extra_memleak_%d_%d_%d", gid, i, r.Next())
		us[i] = NewUnit(key)
	}
	us.Del(c, false)
	for _, u := range us {
		u.Lpush(c, fmt.Sprintf("val_%d", r.Next()))
		ops.Incr()
	}
	us.Del(c, false)
}
