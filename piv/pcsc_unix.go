// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build darwin linux freebsd

package piv

import (
	"fmt"

	pcsc "github.com/gballet/go-libpcsclite"
)

const rcSuccess = pcsc.SCardSuccess

type scContext struct {
	ctx *pcsc.Client
}

func newSCContext() (ctx *scContext, err error) {
	client, err := pcsc.EstablishContext(pcsc.PCSCDSockName, pcsc.ScopeSystem)
	if err != nil {
		return ctx, fmt.Errorf("Error establishing context: %v", err)
	}
	return &scContext{ctx: client}, nil
}

func (c *scContext) Close() error {
	// return scCheck(C.SCardReleaseContext(c.ctx))
	return c.ctx.ReleaseContext()
}

func (c *scContext) ListReaders() (cards []string, err error) {
	return c.ctx.ListReaders()
}

type scHandle struct {
	card *pcsc.Card
}

func (c *scContext) Connect(reader string) (*scHandle, error) {
	var hh scHandle
	var err error
	hh.card, err = c.ctx.Connect(reader, pcsc.ShareExclusive, pcsc.ProtocolT1)
	return &hh, err
}

func (h *scHandle) Close() error {
	return h.card.Disconnect(pcsc.LeaveCard)
}

type scTx struct {
	card *pcsc.Card
}

func (h *scHandle) Begin() (*scTx, error) {
	return &scTx{card: h.card}, nil
}

func (t *scTx) Close() error {
	return t.card.Disconnect(pcsc.LeaveCard)
}

func (t *scTx) transmit(req []byte) (more bool, b []byte, err error) {
	resp, t2, err := t.card.Transmit(req)
	respN := len(resp)
	sw1 := resp[respN-2]
	sw2 := resp[respN-1]
	_, _ = sw1, sw2
	if sw1 == 0x90 && sw2 == 0x00 {
		return false, resp[:respN-2], err
	} else if sw1 == 0x61 {
		return true, resp[:respN-2], nil
	}
	_ = t2
	return false, nil, &apduErr{sw1, sw2}
}
