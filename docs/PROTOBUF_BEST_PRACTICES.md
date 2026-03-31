## Protobuf best practices (exarp-go)

This repo follows the official protobuf “dos/don’ts” guidance:
[Proto Best Practices (Dos and Don’ts)](https://protobuf.dev/best-practices/dos-donts/).
For a general refresher on protobuf’s guarantees and limits, see:
[Overview | Protocol Buffers Documentation](https://protobuf.dev/overview/).

These rules matter because clients and servers roll out asynchronously; wire compatibility is a
durability constraint, not a “nice to have.”

### Schema evolution rules (non-negotiable)

- **Never reuse a field tag number**.
- **Reserve tag numbers** (and ideally names) for deleted fields.
- **Don’t change a field’s type** (unless you’re in one of the explicitly safe cases).
- **Don’t change field semantics via default changes** (proto3 removed explicit defaults; still relevant for “meaning”).
- **Don’t change repeated ↔ scalar** unless you accept lossiness.

### Updating schemas safely (editions guide)

Reference: [Language Guide (editions) — Updating a message type](https://protobuf.dev/programming-guides/editions/#updating).

- **Wire-unsafe changes** (don’t do these on live/interchange/storage schemas):
  - Renumbering fields (changing a tag number).
  - Moving fields into an existing `oneof` (or splitting/merging `oneof`s) without a carefully managed rollout.
- **Wire-safe changes** (safe for binary wire format, subject to app-level API breaks):
  - Adding new fields (old readers ignore; new readers default missing values).
  - Removing fields (but reserve numbers/names; old writers may still send them).
  - Adding enum values (be mindful of exhaustive switches in app code).
- **Conditionally compatible changes** (binary-compatible but can be lossy unless rollout is controlled):
  - Some numeric type changes (e.g. `int32` ↔ `int64`), `enum` ↔ integer types.
  - `string` ↔ `bytes` only when bytes remain valid UTF-8.

### Unknown fields (retention pitfalls)

Unknown fields are preserved in binary, but can be lost when:

- Converting to JSON / text formats.
- Copying messages field-by-field instead of `MergeFrom`/`CopyFrom` style APIs.

### Enums

- Include an **UNSPECIFIED = 0** value as the first enum value.
- Reserve numbers for deleted enum values.
- Prefer enum over bool when the state space might expand.

### API vs storage (design rule)

- Prefer **different messages for RPC APIs vs long-term storage**.
  - If we reuse a message for both, treat it as a “storage schema”: extremely conservative evolution.

### Interchange format (operational rule)

- Avoid relying on **JSON/text proto** interchange for compatibility-sensitive paths.
  - Text/JSON encodings make renames and additions more fragile than binary.

### Operational constraints (from the protobuf overview)

- **Keep messages “small”**: protobuf assumes messages can be loaded into memory and are usually
  “up to a few megabytes.” For larger payloads, consider chunking/streaming or alternative formats.
- **Serialization is not canonical**: the same logical message can have multiple binary encodings.
  Don’t compare serialized bytes for equality or use them as stable cache keys.
- **Forward/backward compatibility is the point**: old code ignores unknown fields; new code reads
  old messages and sees defaults for missing fields. That only holds if we evolve schemas safely.

### Practical checklist for PR review (proto changes)

When reviewing a `.proto` change:

- **Field numbers**:
  - Are new fields using **new** tag numbers (not reused)?
  - Are removed fields/enum values **reserved**?
- **Compatibility**:
  - Did any field type change?
  - Did any field move between `repeated` and scalar?
- **Enums**:
  - Does the enum have `*_UNSPECIFIED = 0`?
- **Surface design**:
  - Is this message for “wire API” or “storage”? Should it be split?

### Local repo notes

- Protos live in `proto/` and are regenerated with `make proto`.
- When adding new “API surface” fields, prefer keeping them **optional** and adding any
  “required-by-contract” semantics in comments or higher-level validation.

