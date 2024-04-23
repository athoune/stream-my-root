package blocks

import "log/slog"

type Tainter struct {
	set    map[string]interface{}
	logger *slog.Logger
}

func NewTainter() *Tainter {
	return &Tainter{
		set:    make(map[string]interface{}),
		logger: slog.Default(),
	}
}

func (t *Tainter) Taint(chunk string) {
	_, ok := t.set[chunk]
	if !ok {
		t.set[chunk] = new(interface{})
		t.logger.Info("New chunk read", "chunk", chunk)
	}
}
