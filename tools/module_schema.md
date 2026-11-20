# Bounded schema adapter

module_schema.py reads one application schema of at most 16 KiB, in addition
to the inherited validatorset schema. It accepts bounded simple fields and
typed unary services, creates signer-annotated descriptors, reflection message
bindings and service adapters, and checks committed parity by default. The
write argument regenerates those files. All compression uses a fixed timestamp
and OS byte. SDK routing and codec round trips are Go scenarios.
