import contextlib
import io
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from edn_format import Keyword, dumps

from codescene_verdict import main, validate


class CodeSceneVerdictTests(unittest.TestCase):
    def setUp(self):
        self.sha = "a" * 40
        self.payload = {
            Keyword("project-id"): "66419",
            Keyword("repo_url"): "https://github.com/loveholidays/mcp-html-share",
            Keyword("commit_sha"): self.sha,
            Keyword("base_ref"): "main",
            Keyword("coverage_results"): {
                Keyword("all_gates_pass"): True,
                Keyword("result_per_gate"): [
                    {Keyword("enabled"): True, Keyword("pass"): True}
                ],
            },
        }

    def output(self):
        prefix = "2026-11-08T12:00:00Z runner"
        logger = "[codescene.devtools.commands.code-coverage-check.backend:102]"
        return (
            f"{prefix} DEBUG {logger} - Gate Coverage Result Payload: {dumps(self.payload)}\n"
            f'{prefix} INFO {logger} - Post result: {{"ok":"ok"}}\n'
        )

    def test_current_passing_verdict_is_accepted(self):
        validate(self.output(), self.sha, "other")

    def test_stale_head_is_rejected(self):
        with self.assertRaisesRegex(ValueError, "commit_sha"):
            validate(self.output(), "b" * 40, "main")

    def test_wrong_project_is_rejected(self):
        self.payload[Keyword("project-id")] = "60749"
        with self.assertRaisesRegex(ValueError, "project-id"):
            validate(self.output(), self.sha, "main")

    def test_wrong_repository_is_rejected(self):
        self.payload[Keyword("repo_url")] = "https://github.com/loveholidays/boost"
        with self.assertRaisesRegex(ValueError, "repo_url"):
            validate(self.output(), self.sha, "main")

    def test_wrong_base_is_rejected(self):
        with self.assertRaisesRegex(ValueError, "base_ref"):
            validate(self.output(), self.sha, "other")

    def test_current_policy_failure_is_rejected(self):
        self.payload[Keyword("coverage_results")][Keyword("all_gates_pass")] = False
        with self.assertRaisesRegex(ValueError, "did not pass"):
            validate(self.output(), self.sha, "main")

    def test_enabled_failure_cannot_hide_behind_overall_pass(self):
        gate = self.payload[Keyword("coverage_results")][Keyword("result_per_gate")][0]
        gate[Keyword("pass")] = False
        with self.assertRaisesRegex(ValueError, "gate failed"):
            validate(self.output(), self.sha, "main")

    def test_no_enabled_gate_is_not_a_pass(self):
        gate = self.payload[Keyword("coverage_results")][Keyword("result_per_gate")][0]
        gate[Keyword("enabled")] = False
        with self.assertRaisesRegex(ValueError, "did not pass"):
            validate(self.output(), self.sha, "main")

    def test_ambiguous_current_verdict_is_rejected(self):
        with self.assertRaisesRegex(ValueError, "Expected one"):
            validate(self.output() + self.output(), self.sha, "main")

    def test_unacknowledged_verdict_is_rejected(self):
        output = self.output().replace('{"ok":"ok"}', '{"error":"unauthorized"}')
        with self.assertRaisesRegex(ValueError, "did not accept"):
            validate(output, self.sha, "main")

    def test_acknowledgement_before_verdict_is_rejected(self):
        lines = self.output().splitlines()
        with self.assertRaisesRegex(ValueError, "acknowledge"):
            validate("\n".join(reversed(lines)), self.sha, "main")

    def test_review_publication_warning_keeps_a_verified_policy_pass(self):
        output = self.output().replace(
            '{"ok":"ok"}',
            '{"ok":"Couldn\'t post review, provider returned clj-http: status 500"}',
        )
        with contextlib.redirect_stdout(io.StringIO()) as messages:
            validate(output, self.sha, "main")
        self.assertIn("::warning::", messages.getvalue())

    def command(self, output):
        with tempfile.TemporaryDirectory() as directory:
            log = Path(directory) / "vendor.log"
            log.write_text(output)
            arguments = ["codescene_verdict.py", str(log), self.sha, "main"]
            with patch("sys.argv", arguments), contextlib.redirect_stdout(io.StringIO()):
                return main()

    def test_cli_accepts_current_vendor_log(self):
        self.assertEqual(0, self.command(self.output()))

    def test_cli_fails_on_malformed_vendor_log(self):
        with contextlib.redirect_stderr(io.StringIO()) as messages:
            self.assertEqual(1, self.command("no verdict"))
        self.assertIn("rejected", messages.getvalue())


if __name__ == "__main__":
    unittest.main()
