# Descriptor normalization

The validator descriptor is decompressed with a 2 MiB expansion cap and a 64 KiB
compressed cap, recompressed with a fixed timestamp and OS byte, and rendered
with stable byte grouping. The schema and wire fields do not change. This
normalization allows the Go module rename to be compared against fresh pinned
generator output without relying on gzip implementation metadata.
