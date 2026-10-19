# Platform ADR 02: exact money

Status: accepted.

Amounts are parsed as decimal text into signed integer USD cents; float conversion and implicit rounding are avoided. USD scale is two fractional digits and is validated at input. Broader multi-currency support is not implemented. See [ADR 002](002-exact-money-and-replay.md).
