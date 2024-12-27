// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestGenerator(t *testing.T) {
	ch := make(chan int64, 100)
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				ch <- Generator.GenerateID()
			}
		}()
	}

	m := make(map[int64]struct{})
	go func() {
		for id := range ch {
			_, ok := m[id]
			fmt.Println(id)
			assert.False(t, ok)
		}
	}()

	wg.Wait()
}
