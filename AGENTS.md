## Agent skills

### Issue tracker

GitHub issues (via the `gh` CLI). See `docs/agents/issue-tracker.md`.

### Triage labels

Default label strings matching the canonical roles. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context (one CONTEXT.md at the repo root). See `docs/agents/domain.md`.

## Development Rules

### Backend Development (TDD)
- **TDD is Mandatory for Backend Changes:** You MUST follow the Test-Driven Development (TDD) red-green-refactor loop for all backend modifications. 
- Write failing tests first to verify architecture, module boundaries, and edge cases before writing implementation code.
- Only after establishing robust, deep, and thoroughly tested module abstractions on the backend should you proceed to build the corresponding frontend views on top of them. End-to-end features must always start with a tested backend foundation.

### Frontend Testing
- The frontend uses `vitest` for unit tests.
- Vitest globals are not configured in `tsconfig.json`. When writing tests (e.g. `*.test.tsx`), you **MUST** explicitly import all testing globals (like `describe`, `it`, `test`, `expect`, `vi`) directly from `'vitest'`. Failure to do so will break the `tsc` build step during Wails compilation.

### Compilation Checks
- **Run Wails Build:** Whenever you make structural changes, add bindings, or modify the backend, you MUST run `wails build` in the project root to proactively catch compilation errors, unused imports, and binding issues before handing the work back to the user.
