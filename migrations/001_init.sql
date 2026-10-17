CREATE TABLE risk_decisions (
    id UUID PRIMARY KEY,
    event_id TEXT NOT NULL,
    score INT NOT NULL,
    decision TEXT NOT NULL,
    reason_codes TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_decisions_event ON risk_decisions(event_id);
