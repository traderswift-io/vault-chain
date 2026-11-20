import gzip
import re
import zlib


def normalize_descriptor(source):
    pattern = r'var fileDescriptor_\w+ = \[\]byte\{\n.*?\n\}'
    match = re.search(pattern, source, re.S)
    if match is None:
        return source
    compressed = bytes(int(value, 16) for value in re.findall(r'0x([0-9a-f]{2})', match[0]))
    if len(compressed) > 65536:
        raise ValueError('descriptor compressed byte limit')
    decoder = zlib.decompressobj(31)
    descriptor = decoder.decompress(compressed, 2097153)
    if len(descriptor) > 2097152 or not decoder.eof or decoder.unused_data:
        raise ValueError('descriptor expanded byte limit')
    data = gzip.compress(descriptor, mtime=0)
    data = data[:9] + bytes([255]) + data[10:]
    body = match[0].split('\n')[0] + '\n'
    body += f'\t// {len(data)} bytes of a gzipped FileDescriptorProto\n'
    body += ''.join('\t' + ''.join(f'0x{byte:02x}, ' for byte in data[index:index + 16]).rstrip() + '\n'
                    for index in range(0, len(data), 16)) + '}'
    return source[:match.start()] + body + source[match.end():]
