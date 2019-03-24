package group

import (
	"context"
	"sync"
)

type Group struct {
	wg      sync.WaitGroup
	errOnce sync.Once
	ctx     context.Context
	err     error
	cancel  func()
}

func NewGroup(ctx context.Context) *Group {
	ctx, cancel := context.WithCancel(ctx)

	return &Group{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (g *Group) Go(fn func(ctx context.Context) error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		if err := fn(g.ctx); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				g.cancel()
			})
		}
	}()
}

func (g *Group) Wait() error {
	g.wg.Wait()
	g.cancel()
	return g.err
}
