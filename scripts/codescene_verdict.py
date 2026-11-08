import json
import re
import sys
from pathlib import Path

from edn_format import Keyword, loads


def vendor_record(output, level, marker):
    pattern = re.compile(
        rf"^\S+\s+\S+\s+{level}\s+"
        rf"\[codescene\.devtools\.commands\.code-coverage-check\.backend:\d+\]"
        rf"\s+-\s+{re.escape(marker)}:\s*(.*)$"
    )
    records = [
        (index, match.group(1))
        for index, line in enumerate(output.splitlines())
        if (match := pattern.fullmatch(line))
    ]
    if len(records) != 1:
        raise ValueError(f"Expected one current {marker} record")
    return records[0]


def validate_context(payload, sha, base):
    expected = {
        "project-id": "66419",
        "repo_url": "https://github.com/loveholidays/mcp-html-share",
        "commit_sha": sha,
        "base_ref": base,
    }
    for key, value in expected.items():
        if payload.get(Keyword(key)) != value:
            raise ValueError(f"Current coverage verdict has the wrong {key}")


def validate_policy(payload):
    results = payload[Keyword("coverage_results")]
    gates = results[Keyword("result_per_gate")]
    enabled = [gate for gate in gates if gate[Keyword("enabled")] is True]
    if not enabled or results[Keyword("all_gates_pass")] is not True:
        raise ValueError("Current CodeScene coverage policy did not pass")
    if any(gate[Keyword("pass")] is not True for gate in enabled):
        raise ValueError("An enabled CodeScene coverage gate failed")


def validate(output, sha, base):
    payload_index, payload = vendor_record(output, "DEBUG", "Gate Coverage Result Payload")
    response_index, response = vendor_record(output, "INFO", "Post result")
    if response_index <= payload_index:
        raise ValueError("CodeScene did not acknowledge the current verdict")
    acknowledgement = json.loads(response)
    allowed = [
        {"ok": "ok"},
        {"ok": "Couldn't post review, provider returned clj-http: status 500"},
    ]
    if acknowledgement not in allowed:
        raise ValueError("CodeScene did not accept the current coverage result")
    parsed = loads(payload)
    validate_context(parsed, sha, base)
    validate_policy(parsed)
    if acknowledgement != allowed[0]:
        print("::warning::CodeScene accepted the verdict but could not publish its review")


def main():
    try:
        log, sha, base = sys.argv[1:]
        validate(Path(log).read_text(), sha, base)
    except (ValueError, KeyError, TypeError, AttributeError, OSError) as error:
        print(f"CodeScene coverage verdict rejected: {error}", file=sys.stderr)
        return 1
    print("Current project, repository, head and base coverage verdict passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
