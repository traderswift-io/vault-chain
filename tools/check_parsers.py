import re
import sys
from collections import Counter
from pathlib import Path


def decoder_sites(root):
    sites = Counter()
    expression = re.compile(r'\b([A-Za-z_]\w*(?:\.[A-Za-z_]\w*)*)\s*(?:\[[^\]\n]+\])?\s*\(')
    decoder = re.compile(r'(?:[Dd]ecode|[Uu]nmarshal|[Pp]arse|[Rr]esolve|CollValue|NewItem|NewMap|NewTxConfig|RegisterFile|RegisterType|Marshal|Encode|AccAddressFromBech32|SetString|Uvarint|ReadFrom|ReadAll|Token|ed25519\.Verify)')
    paths = sorted(path for path in root.rglob('*') if path.suffix in {'.go', '.py'} and not any(part in {'.git', 'build', '_build', '__pycache__'} for part in path.parts))
    if len(paths) > 512:
        raise ValueError('source file count exceeds 512')
    for path in paths:
        if any(part in {'.git', 'build', '_build'} for part in path.parts):
            continue
        if path.stat().st_size > 2097152:
            raise ValueError('source file exceeds 2097152 bytes')
        source = path.read_text()
        covered = []
        for match in expression.finditer(source):
            name = match.group(1)
            python_decoder = path.suffix == '.py' and re.search(r'(read_text|read_bytes|loads|load|decode|parse|compile|finditer|findall|match|search|split|sub|compress|decompress)$', name)
            if decoder.search(name) or python_decoder or name == 'dec':
                covered.append((match.start(1), match.end(1)))
                sites[(path.relative_to(root).as_posix(), name)] += 1
        reference_expression = re.compile(r'\b(?:[A-Za-z_]\w*\.)+[A-Za-z_]\w*\b|\.\s*(?:Unmarshal\w*|Decode\w*)\b')
        for match in reference_expression.finditer(source):
            name = match.group(0).strip('.')
            if decoder.search(name) and not any(start <= match.start() and match.end() <= end for start, end in covered):
                sites[(path.relative_to(root).as_posix(), name)] += 1
    return sites


def check(root):
    declarations = {}
    failures = []
    inventory = root / 'PARSER-INVENTORY.md'
    if inventory.stat().st_size > 2097152:
        print('FAIL check-parsers: inventory exceeds 2097152 bytes')
        return 1
    for line in inventory.read_text().splitlines():
        if not line.startswith('| `'):
            continue
        cells = [cell.strip().strip('`') for cell in line.strip('|').split('|')]
        if len(cells) != 11:
            failures.append('inventory row must have 11 columns: ' + line)
            continue
        path, name, uses, version, byte, depth, count, allocation, work, enforcement, rejection_fuzz = cells
        key = (path, name)
        if key in declarations:
            failures.append('duplicate declaration: ' + str(key))
        try:
            values = [int(uses)]
            for value in (byte, depth, count, allocation, work):
                if value == 'SDK' and 'SDK' in enforcement:
                    continue
                values.append(int(value))
            if min(values) < 1 or not all((version, enforcement, rejection_fuzz)):
                raise ValueError
        except ValueError:
            failures.append('missing or invalid numeric limit: ' + str(key))
            continue
        declarations[key] = values[0]
    actual = decoder_sites(root)
    for key in sorted(actual.keys() | declarations.keys()):
        if actual[key] != declarations.get(key, 0):
            failures.append(f'undeclared decoder or stale count: {key[0]}:{key[1]} actual={actual[key]} declared={declarations.get(key, 0)}')
    if failures:
        print('\n'.join('FAIL check-parsers: ' + failure for failure in failures))
        return 1
    print(f'PASS check-parsers: {sum(actual.values())} sites, {len(declarations)} declarations')
    return 0


if __name__ == '__main__':
    sys.exit(check(Path(sys.argv[1]) if len(sys.argv) > 1 else Path(__file__).resolve().parent.parent))
