# Account-window velocity and transactional replay

Move event, explanation, rules and feedback state behind a transactional repository port. Compute velocity from unique committed events for the same account in a clock-injected 60-second window. Preserve recent_count as a legacy upper signal, never permit it to suppress internal velocity. Memory transactions roll back on errors/cancellation and defensively copy reasons. Durable adapters remain a subsequent card.
