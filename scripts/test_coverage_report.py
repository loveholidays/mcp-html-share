import contextlib
import io
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from coverage_report import (
    CoverageError,
    _inventory,
    _lcov_records,
    _lcov_sources,
    main,
    _package_sources,
    _profile_source,
    _profile_sources,
    _require_complete,
    _run,
)


class CoverageReportTests(unittest.TestCase):
    def test_unimported_production_source_at_zero_stays_in_profile_and_lcov(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / "cmd" / "main.go"
            source.parent.mkdir()
            source.write_text("package main\n", encoding="utf-8")
            profile = root / "coverage.out"
            profile.write_text(
                "mode: atomic\nexample.test/app/cmd/main.go:1.1,1.2 1 0\n",
                encoding="utf-8",
            )
            lcov = root / "coverage.info"
            lcov.write_text(
                f"TN:\nSF:{source}\nDA:1,0\nLF:1\nLH:0\nend_of_record\n",
                encoding="utf-8",
            )
            sources = _profile_sources(profile, "example.test/app")
            converted = _lcov_sources(lcov, root)
            self.assertEqual({"cmd/main.go"}, converted)
            self.assertIn("SF:cmd/main.go", lcov.read_text(encoding="utf-8"))
            _require_complete({"cmd/main.go"}, sources, "Go profile")
            _require_complete({"cmd/main.go"}, converted, "LCOV report")

    def test_profile_omission_fails(self):
        with tempfile.TemporaryDirectory() as directory:
            profile = Path(directory) / "coverage.out"
            profile.write_text("mode: atomic\n", encoding="utf-8")
            with self.assertRaisesRegex(CoverageError, "no production sources"):
                _profile_sources(profile, "example.test/app")

    def test_duplicate_lcov_source_fails(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / "main.go"
            source.write_text("package main\n", encoding="utf-8")
            lcov = root / "coverage.info"
            record = f"SF:{source}\nDA:1,0\nLF:1\nLH:0\nend_of_record\n"
            lcov.write_text(record + record, encoding="utf-8")
            with self.assertRaisesRegex(CoverageError, "duplicate source files"):
                _lcov_sources(lcov, root)

    def test_source_outside_repository_fails(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory) / "repo"
            root.mkdir()
            outside = Path(directory) / "outside.go"
            outside.write_text("package main\n", encoding="utf-8")
            lcov = root / "coverage.info"
            lcov.write_text(f"SF:{outside}\nend_of_record\n", encoding="utf-8")
            with self.assertRaisesRegex(CoverageError, "outside the repository"):
                _lcov_sources(lcov, root)

    def test_invalid_profile_sources_fail(self):
        invalid = [
            "outside.test/app/main.go:1.1,1.2 1 0",
            "example.test/app/../main.go:1.1,1.2 1 0",
            "example.test/app/main.txt:1.1,1.2 1 0",
        ]
        for line in invalid:
            with self.subTest(line=line), self.assertRaises(CoverageError):
                _profile_source(line, "example.test/app")

    def test_malformed_lcov_records_fail(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / "main.go"
            source.write_text("package main\n", encoding="utf-8")
            non_go = root / "readme.txt"
            non_go.write_text("source\n", encoding="utf-8")
            invalid = [
                ([], "no production sources"),
                ([f"SF:{source}", f"SF:{source}"], "one source file"),
                ([f"SF:{non_go}"], "not a Go source"),
                ([f"SF:{source}", "LF:1"], "no executable lines"),
                ([f"SF:{source}", "DA:1,x", "LF:1"], "invalid execution count"),
                ([f"SF:{source}", "DA:1,0", "LF:x"], "invalid line total"),
                ([f"SF:{source}", "DA:1,0", "LF:2"], "no executable lines"),
            ]
            for record, message in invalid:
                with self.subTest(record=record), self.assertRaisesRegex(CoverageError, message):
                    _lcov_sources_from_record(record, root)

    def test_lcov_without_records_or_terminator_fails(self):
        with self.assertRaisesRegex(CoverageError, "missing end_of_record"):
            _lcov_records("SF:main.go")
        with tempfile.TemporaryDirectory() as directory:
            report = Path(directory) / "coverage.info"
            report.write_text("", encoding="utf-8")
            with self.assertRaisesRegex(CoverageError, "no production sources"):
                _lcov_sources(report, directory)

    def test_inventory_and_command_failures_are_reported(self):
        with patch("coverage_report._run", side_effect=["example.test/app", ""]):
            with self.assertRaisesRegex(CoverageError, "inventory is empty"):
                _inventory(Path("."))
        with patch("coverage_report.subprocess.run", return_value=_failed_process()):
            with self.assertRaisesRegex(CoverageError, "go list failed"):
                _run(["go", "list"], Path("."))

    def test_package_source_paths_are_scoped_to_the_module(self):
        package = {
            "ImportPath": "example.test/app/internal",
            "GoFiles": ["service.go"],
            "CgoFiles": ["bridge.go"],
        }
        self.assertEqual(
            {"internal/service.go", "internal/bridge.go"},
            _package_sources(package, "example.test/app"),
        )
        package["GoFiles"] = ["../escape.go"]
        with self.assertRaisesRegex(CoverageError, "invalid production source"):
            _package_sources(package, "example.test/app")

    def test_scope_mismatch_fails(self):
        with self.assertRaisesRegex(CoverageError, "missing expected.go"):
            _require_complete({"expected.go"}, set(), "report")
        with self.assertRaisesRegex(CoverageError, "unexpected actual.go"):
            _require_complete(set(), {"actual.go"}, "report")

    def test_report_error_exits_nonzero(self):
        with patch("coverage_report.validate", side_effect=CoverageError("incomplete")):
            with contextlib.redirect_stderr(io.StringIO()):
                self.assertEqual(1, main())


def _lcov_sources_from_record(record, root):
    report = root / "coverage.info"
    report.write_text("\n".join(record) + "\nend_of_record\n", encoding="utf-8")
    return _lcov_sources(report, root)


def _failed_process():
    class Result:
        returncode = 1
        stderr = "go list failed"

    return Result()


if __name__ == "__main__":
    unittest.main()
