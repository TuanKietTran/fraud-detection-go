# Complete existing success response contracts

Specify GET explanation and GET rules success bodies without changing routes, casing, status codes or behavior. Explanation uses the canonical stored DENY decision, while POST score retains documented BLOCK compatibility. GET rules retains its existing PascalCase RuleConfig fields, exact integer thresholds and bounded weights.

Parse schema/response references and validate actual handler responses against the supported object/array/string/integer/enum/required/bounds subset using existing yaml.v3 and stdlib JSON numbers (no float money conversion). Exercise every declared v1 operation/status, empty202 feedback bodies, learned and extreme valid rule configurations, error request-ID consistency, and negative payload mutations so tests prove rejection rather than only accepting fixtures. This is targeted response-contract verification, not a full OpenAPI semantic validator or runtime policy administration.
