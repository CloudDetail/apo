// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/capitalone/fpe/ff1"
)

const (
	epoch        = int64(1726588800000)
	sequenceBits = 5
	maxSequence  = -1 ^ (-1 << sequenceBits)
	timeShift    = sequenceBits
)

type IDGenerator struct {
	mu       sync.Mutex
	lastTime int64
	sequence int64
}

var Generator = &IDGenerator{}

func (g *IDGenerator) GenerateID() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli()

	if now == g.lastTime {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			for now <= g.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastTime = now

	id := ((now - epoch) << timeShift) | g.sequence
	return id
}

var (
	fpeCipher ff1.Cipher
	fpeDomain = 10
	fpeLength = 14
)

func init() {
	key := os.Getenv("APO_FPE_KEY")
	if key == "" {
		key = config.Get().Server.APOFpeKey
	}

	var err error
	fpeCipher, err = ff1.NewCipher(fpeDomain, fpeLength, []byte(key), nil)
	if err != nil {
		panic(fmt.Errorf("failed to create cipher: %w", err))
	}
}

func (g *IDGenerator) GenerateEncryptedID() (int64, error) {
	raw := g.GenerateID()
	src := fmt.Sprintf("%0*d", fpeLength, raw)
	dstBytes, err := fpeCipher.Encrypt(src)
	if err != nil {
		return 0, err
	}
	cipherStr := string(dstBytes)
	cipherInt, err := strconv.ParseInt(cipherStr, 10, 64)
	if err != nil {
		return 0, errors.New("encrypt output length exceed")
	}
	return cipherInt, nil
}
