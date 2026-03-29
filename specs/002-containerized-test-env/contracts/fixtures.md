# Fixtures Contract

Supported scenarios:
- `default`
- `empty`
- `nested`
- `largefiles`

Contract rules:
- Applying a scenario resets the target bucket to a known baseline before populating scenario data.
- Re-running a scenario after interruption must succeed and restore deterministic state.
- `mc` may be used for fixture setup or inspection only, never as the final proof that s3peep behavior works.
