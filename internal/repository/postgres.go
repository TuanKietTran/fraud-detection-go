package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresEvents stores an atomic policy/event snapshot in a single locked row.
// This favors auditable consistency over throughput: all writers serialize and
// each transaction loads the full history. It is not a large-history adapter.
type PostgresEvents struct{ pool *pgxpool.Pool }

func NewPostgresEvents(ctx context.Context, pool *pgxpool.Pool, config domain.RuleConfig) (*PostgresEvents, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool required")
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	_, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS fraud_state (
  singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
  state jsonb NOT NULL
 )`)
	if err != nil {
		return nil, fmt.Errorf("create fraud state: %w", err)
	}
	initial, err := json.Marshal(EventState{Events: map[string]EventRecord{}, Feedback: map[string]string{}, Rules: config})
	if err != nil {
		return nil, err
	}
	if _, err = pool.Exec(ctx, `INSERT INTO fraud_state(singleton,state) VALUES(true,$1) ON CONFLICT(singleton) DO NOTHING`, initial); err != nil {
		return nil, fmt.Errorf("initialize fraud state: %w", err)
	}
	return &PostgresEvents{pool: pool}, nil
}

func (p *PostgresEvents) Update(ctx context.Context, fn func(*EventState) error) error {
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(rollbackCtx)
	}()
	var raw []byte
	if err = tx.QueryRow(ctx, `SELECT state FROM fraud_state WHERE singleton=true FOR UPDATE`).Scan(&raw); err != nil {
		return err
	}
	var state EventState
	if err = json.Unmarshal(raw, &state); err != nil {
		return fmt.Errorf("decode fraud state: %w", err)
	}
	if state.Events == nil || state.Feedback == nil {
		return fmt.Errorf("invalid stored fraud state")
	}
	if err = fn(&state); err != nil {
		return err
	}
	if err = state.Rules.Validate(); err != nil {
		return err
	}
	updated, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE fraud_state SET state=$1 WHERE singleton=true`, updated); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
