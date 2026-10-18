# HTTP error boundaries

Return consistent JSON errors carrying a server-generated request ID, with bounded 1 MiB bodies and strict single-value JSON decoding on both write routes. Map repository failures to 503, cancellation/deadlines to 408, and recover panics without reflecting sensitive details. Preserve documented legacy 400 replay/feedback conflict behavior to avoid silently changing consumers.

Use request contexts for explanations, rules and feedback so store failure cannot become an empty successful response. Runtime health/readiness, metrics and full server lifecycle remain separate mandatory follow-up work.
