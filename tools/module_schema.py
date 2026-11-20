import gzip
import re
import sys
from pathlib import Path


def varint(value):
    result = bytearray()
    while value > 127:
        result.append((value & 127) | 128)
        value >>= 7
    result.append(value)
    return bytes(result)


def field(number, value):
    if isinstance(value, int):
        return varint(number * 8) + varint(value)
    if isinstance(value, str):
        value = value.encode()
    return varint(number * 8 + 2) + varint(len(value)) + value


def camel(name):
    return ''.join(part.capitalize() for part in name.split('_'))


def parse_schema(source):
    if len(source) > 16384:
        raise ValueError('schema byte limit')
    namespace = re.search(r'^package ([\w.]+);$', source, re.M).group(1)
    package = re.search(r'^option go_package = "([\w/-]+)";$', source, re.M).group(1)
    messages = []
    for name, body in re.findall(r'message (\w+) \{([^}]*)\}', source):
        signer = re.findall(r'  option \(cosmos.msg.v1.signer\) = "(\w+)";\n', body)
        rows = re.findall(r'  (repeated )?(\w+) (\w+) = (\d+);\n', body)
        expected = ('\n' if body else '')
        expected += ''.join(f'  option (cosmos.msg.v1.signer) = "{value}";\n' for value in signer)
        expected += ''.join(f'  {repeated}{kind} {field_name} = {number};\n' for repeated, kind, field_name, number in rows)
        if body != expected or [int(row[3]) for row in rows] != list(range(1, len(rows) + 1)):
            raise ValueError('unsupported message schema')
        messages.append((name, signer, rows))
    services = []
    for name, body in re.findall(r'service (\w+) \{([^}]*)\}', source):
        methods = re.findall(r'  rpc (\w+)\((\w+)\) returns \((\w+)\);\n', body)
        expected = '\n' + ('  option (cosmos.msg.v1.service) = true;\n' if name == 'Msg' else '')
        expected += ''.join(f'  rpc {method}({request}) returns ({response});\n' for method, request, response in methods)
        if body != expected:
            raise ValueError('unsupported service schema')
        services.append((name, methods))
    if len(messages) > 32 or len(services) != 2:
        raise ValueError('schema count limit')
    return namespace, package, messages, services


def descriptor(namespace, package, messages, services):
    module = package.split('/')[-1]
    result = field(1, f'traderswift/{module}/v1/{module}.proto') + field(2, namespace)
    result += field(3, 'cosmos/msg/v1/msg.proto')
    kinds = {'string': 9, 'bytes': 12, 'uint64': 4, 'bool': 8}
    for name, signer, rows in messages:
        message = field(1, name)
        for repeated, kind, field_name, number in rows:
            json_name = camel(field_name)
            json_name = json_name[0].lower() + json_name[1:]
            entry = field(1, field_name) + field(3, int(number)) + field(4, 3 if repeated else 1)
            entry += field(5, kinds.get(kind, 11))
            if kind not in kinds:
                entry += field(6, '.' + namespace + '.' + kind)
            message += field(2, entry + field(10, json_name))
        for value in signer:
            message += field(7, field(11110000, value))
        result += field(4, message)
    for name, methods in services:
        service = field(1, name)
        for method, request, response in methods:
            service += field(2, field(1, method) + field(2, '.' + namespace + '.' + request) + field(3, '.' + namespace + '.' + response))
        if name == 'Msg':
            service += field(3, field(11110000, 1))
        result += field(6, service)
    return result + field(8, field(11, package)) + field(12, 'proto3')


def message_source(module, namespace, index, name, rows):
    result = f'package {module}\n\nimport "github.com/cosmos/gogoproto/proto"\n\n'
    fields = []
    for repeated, kind, field_name, number in rows:
        go_name = camel(field_name)
        go_type = {'bytes': '[]byte', 'string': 'string', 'uint64': 'uint64', 'bool': 'bool'}.get(kind, '*' + kind)
        if repeated:
            go_type = '[]' + go_type
        wire = 'varint' if kind in {'uint64', 'bool'} else 'bytes'
        tag = f'{wire},{number},{"rep" if repeated else "opt"},name={field_name}'
        json_name = go_name[0].lower() + go_name[1:]
        if json_name != field_name:
            tag += ',json=' + json_name
        fields.append((go_name, go_type, f'`protobuf:"{tag},proto3" json:"{field_name},omitempty"`'))
    if fields:
        result += f'type {name} struct {{\n'
        width_name = max(len(row[0]) for row in fields)
        width_type = max(len(row[1]) for row in fields)
        result += ''.join(f'\t{go_name:<{width_name}} {go_type:<{width_type}} {tag}\n' for go_name, go_type, tag in fields)
        result += '}\n'
    else:
        result += f'type {name} struct{{}}\n'
    result += f'\nfunc (message *{name}) Reset() {{\n\t*message = {name}{{}}\n}}\n'
    result += f'\nfunc (message *{name}) String() string {{\n\treturn proto.CompactTextString(message)\n}}\n'
    result += f'\nfunc (*{name}) ProtoMessage() {{}}\n'
    result += f'\nfunc (*{name}) Descriptor() ([]byte, []int) {{\n\treturn moduleDescriptor, []int{{{index}}}\n}}\n'
    result += f'\nfunc init() {{\n\tproto.RegisterType((*{name})(nil), "{namespace}.{name}")\n}}\n'
    return result


def service_source(module, namespace, name, methods):
    result = f'package {module}\n\nimport (\n\t"context"\n\n\t"google.golang.org/grpc"\n)\n\n'
    result += f'type {name}Server interface {{\n'
    result += ''.join(f'\t{method}(context.Context, *{request}) (*{response}, error)\n' for method, request, response in methods) + '}\n'
    for method, request, response in methods:
        result += f'\nfunc handle{name}{method}(server interface{{}}, ctx context.Context, dec func(interface{{}}) error, interceptor grpc.UnaryServerInterceptor) (interface{{}}, error) {{\n'
        result += f'\tmessage := &{request}{{}}\n\tif err := dec(message); err != nil {{\n\t\treturn nil, err\n\t}}\n'
        result += f'\tif interceptor == nil {{\n\t\treturn server.({name}Server).{method}(ctx, message)\n\t}}\n'
        result += f'\tinfo := &grpc.UnaryServerInfo{{Server: server, FullMethod: "/{namespace}.{name}/{method}"}}\n'
        result += f'\thandler := func(ctx context.Context, request interface{{}}) (interface{{}}, error) {{\n\t\treturn server.({name}Server).{method}(ctx, request.(*{request}))\n\t}}\n'
        result += '\treturn interceptor(ctx, message, info, handler)\n}\n'
    result += f'\nvar {name}ServiceDesc = grpc.ServiceDesc{{\n\tServiceName: "{namespace}.{name}",\n\tHandlerType: (*{name}Server)(nil),\n\tMethods: []grpc.MethodDesc{{\n'
    result += ''.join(f'\t\t{{MethodName: "{method}", Handler: handle{name}{method}}},\n' for method, _, _ in methods)
    result += f'\t}},\n\tStreams:  []grpc.StreamDesc{{}},\n\tMetadata: "traderswift/{module}/v1/{module}.proto",\n}}\n'
    return result


def main():
    root = Path(__file__).resolve().parent.parent
    paths = [path for path in (root / 'proto').glob('*.proto') if path.stem != 'validatorset']
    if len(paths) != 1 or paths[0].stat().st_size > 16384:
        raise ValueError('expected one bounded application schema')
    namespace, package, messages, services = parse_schema(paths[0].read_text())
    module = package.split('/')[-1]
    expected = {}
    for index, (name, _, rows) in enumerate(messages):
        expected[f'binding_{index:02}.go'] = message_source(module, namespace, index, name, rows)
    for name, methods in services:
        expected[f'binding_{name.lower()}_service.go'] = service_source(module, namespace, name, methods)
    compressed = gzip.compress(descriptor(namespace, package, messages, services), mtime=0)
    compressed = compressed[:9] + bytes([255]) + compressed[10:]
    content = f'package {module}\n\nimport "github.com/cosmos/gogoproto/proto"\n\n'
    content += f'func init() {{\n\tproto.RegisterFile("traderswift/{module}/v1/{module}.proto", moduleDescriptor)\n}}\n\nvar moduleDescriptor = []byte{{\n'
    for offset in range(0, len(compressed), 12):
        content += '\t' + ', '.join(f'0x{byte:02x}' for byte in compressed[offset:offset + 12]) + ',\n'
    expected['binding_descriptor.go'] = content + '}\n'
    writing = len(sys.argv) == 2 and sys.argv[1] == 'write'
    present = {path.name for path in (root / module).glob('binding_*.go')}
    if not writing and present != expected.keys():
        raise ValueError('binding file set differs')
    for name, content in expected.items():
        path = root / module / name
        if writing:
            path.write_text(content)
        elif path.read_text() != content:
            raise ValueError('schema/binding drift: ' + name)
    print('PASS module schema and reflection binding parity')


if __name__ == '__main__':
    main()
