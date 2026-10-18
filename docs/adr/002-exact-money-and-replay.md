# ADR 002: Exact USD cents and immutable replay

Status: accepted.

JSON decimal strings are canonical. A transitional raw numeric token is parsed as text, never through float64. USD is the only currency; at most two fractional digits are accepted, with no implicit rounding. Amounts must fit signed 64-bit minor units and be nonnegative. Unsupported currencies/scales fail explicitly.

The canonical scoring payload includes event identity, account, amount cents and contextual signals. Event identity is the write-idempotency identity. Exact replay returns the original stored decision even after feedback or time changes; changed input with the same identity fails. This keeps retries from contributing to account-window velocity twice. The HTTP score route still emits BLOCK as a legacy alias of DENY; clients must accept both.

Account velocity uses prior unique events in (now-60s, now]. Clock rollback excludes future timestamps rather than treating them as recent. The legacy caller count can increase, but cannot decrease, the internal count. Injected clocks make all boundaries deterministic in tests.

Trade-offs: no foreign-exchange conversion or multi-currency scale registry; no replay-expiry policy; unbounded history needs a retention design. The event ID replaces a separate header for score idempotency, while feedback identity is the event/label pair. No genuine payment-network decision or machine-learning accuracy is claimed.
