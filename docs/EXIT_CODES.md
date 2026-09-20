# Exit codes

json-doctor keeps its exit status intentionally small and script-friendly.

| Code | Meaning |
|---|---|
| 0 | Input was processed successfully. |
| 1 | Invalid JSON, invalid query, or invalid command input. |

## Automation guidance

For CI or pre-commit hooks:
1. pipe or pass the JSON input;
2. check the process exit status;
3. only consume formatted output when the status is zero;
4. use --explain when an invalid document needs a human-readable location.

Keep stdout suitable for successful command output and stderr suitable for diagnostics.
