# SQL Migration Assembly Tool

## Overview

This directory implements a **fragment-based** approach to managing
[goose](https://github.com/pressly/goose) SQL migrations. Instead of
writing one enormous SQL file per migration version, you break the DDL
into small, focused **fragment files** and let the `gen-migration.sh`
script assemble them into a single goose-compatible migration.

### Why fragments?

| Concern | Monolith `.sql` | Fragment approach |
|---------|-----------------|-------------------|
| Readability | Hard to navigate 1000+ line files | Each fragment has a clear scope |
| Code review | Diff noise in a single giant file | Small, isolated diffs |
| Team collaboration | Merge conflicts | Fragments are independent files |
| Ordering control | Manual | Numeric prefix on filenames |

---

## Directory layout

```
migrations/
├── gen-migration.sh          # assembler script
├── README.md                 # this file
├── 0001000_init/             # ← one directory = one migration version
│   ├── 0001000_common.sql    #    fragment: schemas, extensions, domains
│   ├── 0001001_tables.sql    #    fragment: tables & triggers
│   ├── 0001002_views.sql     #    fragment: views & helpers
│   ├── 0001003_listers.sql   #    fragment: list functions
│   └── 0001004_syncers.sql   #    fragment: sync functions
└── 0001000_init.sql          # ← generated output (DO NOT EDIT BY HAND)
```

### Naming conventions

* **Directory name** becomes the output filename:
  `0001000_init/` → `0001000_init.sql`.
* **Fragment files** inside each directory are sorted by their numeric
  prefix (`0001000`, `0001001`, …). The order determines the order of
  SQL statements in the assembled migration.

---

## Fragment file format

Every fragment **must** follow the standard goose annotation structure:

```sql
-- +goose Up
-- +goose StatementBegin

-- Your DDL / DML for the "up" direction goes here.
CREATE TABLE ...;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Your rollback DDL / DML goes here (may be empty).
DROP TABLE IF EXISTS ...;

-- +goose StatementEnd
```

* Content between `-- +goose StatementBegin` and `-- +goose StatementEnd`
  in the **Up** section is collected in ascending file order.
* Content in the **Down** section is collected in **reverse** file order
  (last fragment's Down is applied first), which is the conventional
  rollback strategy — undo the most recent changes first.

---

## Usage

```bash
# From the migrations/ directory
./gen-migration.sh              # process every subdirectory
./gen-migration.sh -d 0001000_init   # process only this directory
./gen-migration.sh -o /tmp/out       # write output to a custom location
./gen-migration.sh -d 0001000_init -o /tmp/out  # combine flags
```

### Flags

| Flag | Argument | Default | Description |
|------|----------|---------|-------------|
| `-d` | `DIR` | *(all subdirs)* | Directory with fragment files. Can be repeated (`-d dir1 -d dir2`). |
| `-o` | `DIR` | script's own directory | Where to write the assembled `.sql` file(s). |
| `-h` | — | — | Print usage and exit. |

### Examples

**1. Generate all migrations (typical CI/CD or build step)**

```bash
cd internal/sg-server/repository/pg/migrations
./gen-migration.sh
# → Generated: .../migrations/0001000_init.sql
```

**2. Regenerate a single migration after editing a fragment**

```bash
./gen-migration.sh -d 0001000_init
```

**3. Output to a staging directory for review**

```bash
./gen-migration.sh -o ./review
# Creates ./review/0001000_init.sql
```

---

## Workflow: adding a new fragment

1. Create a new `.sql` file inside the relevant directory with the next
   numeric prefix:

   ```bash
   touch 0001000_init/0001005_new_feature.sql
   ```

2. Fill it in using the fragment template:

   ```sql
   -- +goose Up
   -- +goose StatementBegin

   CREATE TABLE sgroups.tbl_new_feature ( ... );

   -- +goose StatementEnd

   -- +goose Down
   -- +goose StatementBegin

   DROP TABLE IF EXISTS sgroups.tbl_new_feature;

   -- +goose StatementEnd
   ```

3. Re-run the generator:

   ```bash
   ./gen-migration.sh -d 0001000_init
   ```

4. Commit **both** the new fragment and the regenerated `.sql` file.

---

## Workflow: adding a new migration version

1. Create a new directory with the appropriate version prefix:

   ```bash
   mkdir 0002000_add_policies
   ```

2. Add fragment files inside it (`0002000_xxx.sql`, `0002001_yyy.sql`, …).

3. Run the generator — it will automatically pick up the new directory:

   ```bash
   ./gen-migration.sh
   # → Generated: .../migrations/0002000_add_policies.sql
   ```

---

## Important notes

* **Do not edit the generated `.sql` file by hand.** It will be
  overwritten on the next run of `gen-migration.sh`.
* The Down sections are assembled in **reverse order** relative to the
  fragment file names. This ensures that objects created last are dropped
  first, avoiding dependency errors during rollback.
* Fragment files that contain only empty Up/Down sections are silently
  skipped (no content is emitted for them).
