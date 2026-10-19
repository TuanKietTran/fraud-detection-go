# Platform ADR 03: event identity and idempotency

Status: accepted.

`event_id` is the score-write identity and the canonical payload is stored with the decision. Exact retries return the original result and contribute to account velocity once; a changed payload with the same identity preserves the legacy HTTP400 conflict. Feedback uses event/label identity. See [ADR 002](002-exact-money-and-replay.md).
