# Pinned binding split

split_proto.py preserves the reference gocosmos declaration order and imports,
splitting files below 400 lines. Large Unmarshal cases are moved into helpers
without changing their parsing behavior. Only validatorset uses this path now;
the removed storage schema is not regenerated. module_schema.py separately
checks the new module's small reflection adapters against its protobuf schema.

Always split the complete pinned generator output, never concatenated split
files. Split-file headers and boundary whitespace alter declaration grouping on
a second split. The committed validatorset files were rebuilt from the saved,
identical generator output of both generator rounds; check-proto still regenerates
twice and compares every committed byte.
