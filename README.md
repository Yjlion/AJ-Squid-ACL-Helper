# AJ-Squid-ACL-Helper

`AJ-Squid-ACL-Helper` is now implemented as standalone Go binaries for Squid external ACL checks and ACL list management.

The project provides:
- `aj-helper`: Squid `external_acl_type` helper that checks domain/url/regex ACL entries.
- `aclctl`: command-line tool to add, delete, list, and purge ACL entries.

ACL data is stored as plain text files under `lists/<group>/`:
- `domains`
- `urls`
- `expressions`

## Build

```bash
go build -o bin/aj-helper ./cmd/aj-helper
go build -o bin/aclctl ./cmd/aclctl
```

## Squid integration

```conf
external_acl_type aj_acl %ACL %DST %URI /path/to/bin/aj-helper --base-dir /path/to/repo
acl porn external aj_acl
acl mylist external aj_acl
```

## Manage ACLs

```bash
# add
./bin/aclctl add porn url sexy.com
./bin/aclctl add porn domain bad.example
./bin/aclctl add porn er 'https?://evil\\.example/.*'

# delete
./bin/aclctl del porn url sexy.com

# list
./bin/aclctl list
./bin/aclctl list porn
./bin/aclctl list porn er

# purge
./bin/aclctl purge porn
./bin/aclctl purge porn domain
```

## Releases

Tagged versions (`v*`) automatically publish x64 binaries to GitHub Releases using `.github/workflows/release-x64.yml`.

Release assets include:
- `aj-squid-acl-helper_<version>_linux_amd64.tar.gz`
- `aj-squid-acl-helper_<version>_darwin_amd64.tar.gz`
- `aj-squid-acl-helper_<version>_windows_amd64.zip`
- `checksums.txt` (SHA-256 for all archives)

To cut a release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Notes

- Legacy Perl scripts (`aj_helper.pl`, `acl.pl`) are kept in the repository for reference.
- The Go version does not require BerkeleyDB.
