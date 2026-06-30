# Register map identity and config format v2

Register maps will use generated stable identities for RegisterGroups and RegisterDefinitions, while keeping user-facing names and Modbus addresses editable. We are making a clean break to config format v2, storing groups and definitions as arrays with IDs inside, adding `config_version`, and not providing a migration path for old saved configs.

## Considered Options

- Use group names and definition start addresses as identity. This matched the current model but made renames, address edits, moves, and watches brittle.
- Add stable generated IDs only for groups. This solved group rename but left definitions fragile during address edits and bulk operations.
- Add generated IDs for both groups and definitions. This makes identity independent of user labels and register offsets, at the cost of a breaking config format change.

## Consequences

- Group names are editable and must be unique within a Device, but hidden generated IDs are used internally.
- RegisterDefinition names are labels and do not need to be unique; address ranges still must not overlap within a RegisterGroup.
- Watch-list and polling data can key off stable definition identity rather than mutable register addresses.
- Saved configs use `config_version: 2`; old configs are not migrated as part of this clean break.
