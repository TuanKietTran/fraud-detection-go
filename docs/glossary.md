# Domain glossary

| Term | Meaning in this synthetic service |
| --- | --- |
| Minor unit | One USD cent, represented by int64. No binary floating-point money or implicit rounding. |
| Event | A unique scored attempt, identified by event_id. A denial is still an observed attempt. |
| Replay | Reusing an event identity with the same canonical payload; returns its original decision without another history effect. |
| Payload conflict | Same event identity with changed canonical input; legacy HTTP400, not a transient retry condition. |
| Account window | Prior unique attempts for one account with timestamps in (now-60s,now]; the lower endpoint and future timestamps are excluded. |
| Velocity | Number of qualifying prior attempts, independent of their decisions. Legacy supplied recent_count can only increase this signal. |
| Structuring band | Synthetic amount band; inferred pattern requires three unique active-window attempts in its inclusive top10%. Default USD900..1000. |
| Weighted score | Sum of named rule weights, capped at100. Reasons are retained even when the sum is capped. |
| ALLOW / REVIEW / DENY | Default score bands below40,40..69,70..100. Feedback can change the review threshold. |
| BLOCK | Legacy score-response alias of DENY; stored explanations use canonical DENY. |
| Explanation | Stored event decision and named reasons, not a recomputation under current rules. |
| Rule snapshot | Validated policy used when the event was first scored, retained with that event. |
| Confirmed fraud | Synthetic feedback label lowering ReviewAt by one, minimum1. Not an accuracy claim. |
| False positive | Synthetic feedback label raising ReviewAt by one, maximumDenyAt. Conflicting feedback for an event is rejected. |
| Scoped transaction | Memory view selecting addressed event/feedback, requested account history and global policy; scope escapes fail without partial effects. |
| Durable | SQL state survives an actual PostgreSQL stop/start with retained data directory; this alone does not prove scalable or production recovery. |
| Ambiguous response | A request failed/canceled after its write committed. Retrying the original identity recovers the original result. |
| Readiness | Ability to access the selected event store; separate from process liveness. |
| Weighted coverage | Covered non-cmd statements divided by total non-cmd statements, not the average of package percentages. |
