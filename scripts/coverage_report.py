import json
import os
import subprocess
import sys
from pathlib import Path, PurePosixPath


class CoverageError(ValueError):
    pass


def _run(command, root):
    result = subprocess.run(command, cwd=root, capture_output=True, text=True, check=False)
    if result.returncode:
        raise CoverageError(result.stderr.strip() or "coverage command failed")
    return result.stdout


def _tracked_sources(root):
    tracked = _run(["git", "ls-files", "--", "*.go"], root).splitlines()
    return {path for path in tracked if not path.endswith("_test.go")}


def _inventory(root):
    module = _run(["go", "list", "-m"], root).strip()
    output = _run(["go", "list", "-json", "./..."], root)
    decoder = json.JSONDecoder()
    packages = []
    while output.strip():
        output = output.lstrip()
        package, end = decoder.raw_decode(output)
        packages.append(package)
        output = output[end:]
    files = set()
    for package in packages:
        files.update(_package_sources(package, module))
    if not files:
        raise CoverageError("Go package inventory is empty")
    _require_complete(_tracked_sources(root), files, "Go package inventory")
    return module, files


def _package_sources(package, module):
    directory = package["ImportPath"].removeprefix(module + "/")
    files = package.get("GoFiles", []) + package.get("CgoFiles", [])
    sources = set()
    for name in files:
        path = PurePosixPath(directory, name)
        if path.is_absolute() or ".." in path.parts:
            raise CoverageError(f"invalid production source path: {path}")
        sources.add(path.as_posix())
    return sources


def _valid_go_source(source):
    return not source.is_absolute() and ".." not in source.parts and source.suffix == ".go"


def _profile_source(line, module):
    location = line.split(None, 1)[0].split(":", 1)[0]
    prefix = module + "/"
    if not location.startswith(prefix):
        raise CoverageError(f"coverage source is outside the module: {location}")
    source = PurePosixPath(location[len(prefix):])
    if not _valid_go_source(source):
        raise CoverageError(f"invalid coverage source path: {location}")
    return source.as_posix()


def _profile_sources(path, module):
    sources = {
        _profile_source(line, module)
        for line in path.read_text(encoding="utf-8").splitlines()[1:]
    }
    if not sources:
        raise CoverageError("Go coverage profile has no production sources")
    return sources


def _lcov_source_name(record):
    sources = [line[3:] for line in record if line.startswith("SF:")]
    if len(sources) != 1:
        raise CoverageError("LCOV record must contain one source file")

    return sources[0]


def _repository_relative(source, root):
    source = Path(source).resolve()
    try:
        relative = source.relative_to(root).as_posix()
    except ValueError as error:
        raise CoverageError(f"LCOV source is outside the repository: {source}") from error
    return source, relative


def _repository_source(record, root):
    source, relative = _repository_relative(_lcov_source_name(record), root)
    if not source.is_file() or source.suffix != ".go":
        raise CoverageError(f"LCOV source is not a Go source file: {relative}")
    return relative


def _valid_line_total(record, source):
    totals = [line[3:] for line in record if line.startswith("LF:")]
    if len(totals) != 1 or not totals[0].isdigit():
        raise CoverageError(f"LCOV source has an invalid line total: {source}")
    return int(totals[0])


def _valid_hit_count(line, source):
    parts = line[3:].split(",")
    if len(parts) != 2 or not parts[-1].isdigit():
        raise CoverageError(f"LCOV source has an invalid execution count: {source}")
    return int(parts[-1])


def _validate_line_coverage(record, source):
    counts = [line for line in record if line.startswith("DA:")]
    total = _valid_line_total(record, source)
    if not counts or total != len(counts):
        raise CoverageError(f"LCOV source has no executable lines: {source}")
    for line in counts:
        _valid_hit_count(line, source)


def _lcov_source(record, root):
    source = _repository_source(record, root)
    _validate_line_coverage(record, source)
    return source


def _lcov_records(content):
    chunks = content.split("end_of_record\n")
    if chunks[-1].strip():
        raise CoverageError("LCOV record is missing end_of_record")
    return [chunk.splitlines() for chunk in chunks[:-1] if chunk.strip()]


def _write_normalized_lcov(path, records, sources):
    normalized = [
        [f"SF:{source}" if line.startswith("SF:") else line for line in record]
        for record, source in zip(records, sources, strict=True)
    ]
    content = "".join("\n".join(record) + "\nend_of_record\n" for record in normalized)
    path.write_text(content, encoding="utf-8")


def _lcov_sources(path, root):
    root = Path(root).resolve()
    content = path.read_text(encoding="utf-8")
    records = _lcov_records(content)
    sources = [_lcov_source(record, root) for record in records]
    if len(sources) != len(set(sources)):
        raise CoverageError("LCOV report contains duplicate source files")
    if not sources:
        raise CoverageError("LCOV report has no production sources")
    _write_normalized_lcov(path, records, sources)
    return set(sources)


def _require_complete(expected, actual, report):
    missing = sorted(expected - actual)
    unexpected = sorted(actual - expected)
    if missing or unexpected:
        detail = []
        if missing:
            detail.append("missing " + ", ".join(missing))
        if unexpected:
            detail.append("unexpected " + ", ".join(unexpected))
        raise CoverageError(f"{report} source scope mismatch: {'; '.join(detail)}")


def validate(root, profile, lcov):
    root = Path(root).resolve()
    module, expected = _inventory(root)
    _require_complete(expected, _profile_sources(Path(profile), module), "Go profile")
    _require_complete(expected, _lcov_sources(Path(lcov), root), "LCOV report")
    return len(expected)


def main():
    root = Path(os.environ.get("GITHUB_WORKSPACE", Path(__file__).resolve().parents[1]))
    try:
        count = validate(root, root / "coverage.out", root / "coverage.info")
    except (CoverageError, OSError, json.JSONDecodeError) as error:
        print(f"Coverage report validation failed: {error}", file=sys.stderr)
        return 1
    print(f"Validated coverage for {count} production Go files")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
