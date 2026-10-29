// Package execute submits vault actions on-chain. It is deliberately dumb
// plumbing: because the vault is a bounded executor, every Action is rejected
// on-chain if it breaks caps / idle-floor / churn, so the Executor never
// re-checks those.
package execute

import (
	"context"
	"log/slog"

	"github.com/saintsama98/Gauge-Keeper-Go/pkg/types"
)

// Executor submits a single Action and returns its tx hash.
//
// LocalExecutor (local.go) signs with the curator key and LogExecutor logs
// without sending. Another signer, such as a relayer that holds the key, can
// implement this interface without touching the keeper loop.
type Executor interface {
	Execute(ctx context.Context, action types.Action) (txHash string, err error)
}

// LogExecutor is a dry-run executor: it logs the intended action and returns a
// placeholder hash. Default until a signer + bindings are wired.
type LogExecutor struct{ log *slog.Logger }

func NewLogExecutor(log *slog.Logger) *LogExecutor { return &LogExecutor{log: log} }

func (e *LogExecutor) Execute(ctx context.Context, action types.Action) (string, error) {
	e.log.Info("DRY-RUN action", "kind", action.Kind.String())
	return "0xDRYRUN", nil
}
