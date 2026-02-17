# Codebase Task Proposals

After reviewing `README.md`, `aj_helper.pl`, and `acl.pl`, here are four focused tasks to improve quality.

## 1) Typo fix task
**Title:** Fix wording and grammar typos in README introduction

**Issue found:** The README contains multiple typos/awkward phrases such as `matchs acls` and `This keep centralized config`.

**Proposed task:**
- Rewrite the first two README paragraphs in clear English while preserving meaning.
- Keep the command examples unchanged.

**Acceptance criteria:**
- No occurrences of `matchs` remain.
- Intro text reads naturally and is technically accurate.

---

## 2) Bug fix task
**Title:** Fix `del` command deleting the wrong key in `acl.pl`

**Issue found:** `del()` prefixes the lookup key with a dot (`db_del(".$value", undef)`), while `add()` stores `$value` exactly as provided. This means deleting a value that was added usually fails.

**Proposed task:**
- Change `del()` to call `db_del($value, undef)`.
- Return a clear status message when a key is not found.

**Acceptance criteria:**
- `add` then `del` of the same value removes the key.
- `list` no longer shows deleted entries.

---

## 3) Comment/documentation discrepancy task
**Title:** Align README `purge` examples with actual implementation

**Issue found:** README advertises `purge` usage, but in `acl.pl` the `purge` action only prints `TODO - Apagando todos dados` and performs no purge.

**Proposed task (choose one):**
- Either implement real `purge` behavior, **or**
- Update README/help text to mark `purge` as not implemented.

**Acceptance criteria:**
- Documentation and runtime behavior match.
- Users are not misled about `purge` availability.

---

## 4) Test improvement task
**Title:** Add regression tests for CLI CRUD behavior in `acl.pl`

**Issue found:** There are no automated tests covering the CLI workflow (`add`, `list`, `del`, and not-found behavior).

**Proposed task:**
- Add a lightweight test script (e.g., Perl TAP via `Test::More`) using a temporary test directory.
- Cover at least:
  - `add` writes entries,
  - `list` shows expected keys,
  - `del` removes existing key,
  - deleting a missing key reports correctly.

**Acceptance criteria:**
- Tests run via one documented command.
- Includes regression coverage for the `del` key-format bug.
