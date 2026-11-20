import sys
from pathlib import Path

root = Path(__file__).resolve().parent.parent
failures = []
for path in sorted(root.rglob('*')):
    if any(part in {'.git', 'build', '_build'} for part in path.parts) or not path.is_file():
        continue
    if path.suffix not in {'.go', '.py', '.proto', '.sh'}:
        continue
    if path.stat().st_size > 2097152:
        failures.append(f'{path.relative_to(root)} exceeds source byte limit')
        continue
    if len(path.read_text().splitlines()) >= 400:
        failures.append(f'{path.relative_to(root)} reaches the 400-line cap')
    if 'tests' in path.parts:
        failures.append('committed tests directory is forbidden')
if failures:
    print('\n'.join('FAIL layout: ' + failure for failure in failures))
    sys.exit(1)
print('PASS source layout')
