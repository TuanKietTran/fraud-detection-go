# Definition of done

A box is **Met** only when the code behaves that way and a test asserts it. **Partial** means the code does it, but no test asserts that scenario. **Unmet** means it is not implemented.

Playback storyboard: `openspec/gifs/playback-flow.gif`.

## Met

| Done when | Evidence |
| --- | --- |
| Amount at or under 5000 is ALLOW, score 0 | Playback uses amount 100 and requires an ALLOW body. |
| Amount over 5000 and at or under 10000 is REVIEW, score 40, reason HIGH_AMOUNT | Playback uses 6000 and requires a REVIEW body. |
| Amount over 10000 is BLOCK, score 70, reasons HIGH_AMOUNT and VERY_HIGH_AMOUNT | Playback uses 20000 and requires a BLOCK body. |
| `POST /v1/score` does not persist the decision | The server never calls the repository. Replay returns the same JSON. |
| 1000 scores finish in under 60 seconds, with one HTTP 503 retried on each call | Playback test, 77ms. |
| Replay of the same events returns the same body and HTTP 200 | Playback test. |
| HTTP 400 is not retried | Playback unit test. |
| Statement coverage is measured | 55.9%. |

## Partial

| Done when | Why it is not Met |
| --- | --- |
| Invalid JSON is HTTP 400 | Implemented. No test sends a bad body. |
| Amount exactly 5000 is ALLOW and exactly 10000 is REVIEW | Implemented with strict `>`. The playback amounts are 100, 6000, and 20000, not the boundaries. |

## Unmet

| Not done | Why |
| --- | --- |
| RecordDecision, UpsertRule, GetDecision | The in-memory store exists and the server does not call it. |
| Postgres decisions | `migrations/001_init.sql` is not applied by the server. |
| Velocity or non-amount rules | The score uses amount only. |
| Coverage at a high bar | 55.9%. `cmd/server` and the unused repository are uncovered. No 80% gate. |

## GIF review

Good enough, and it is the one flow that must not look like the others. Playback is HTTP 200 and "same decisions", not HTTP 400, because nothing is stored. Storyboard, not a screen capture.
