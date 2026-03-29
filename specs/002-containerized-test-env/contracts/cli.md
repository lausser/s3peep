# CLI Contract

Covered CLI surfaces for spec 002:
- `s3peep init`
- `s3peep profile add --name --region --access-key --secret [--endpoint] [--bucket]`
- `s3peep profile list`
- `s3peep profile switch --name`
- `s3peep profile remove --name`
- `s3peep serve [--config] [--port]`

Contract rules:
- Success cases must assert exit code `0` and expected stdout or config-file state changes.
- Failure cases must assert non-zero exit codes and specific error output.
- Helper-generated files do not count as proof of CLI behavior unless the CLI reads or writes them as part of the test.
