# CLAUDE.md

Guidance for Claude Code in this repository. Start with [README.md](README.md); build, test and extension steps are in [CONTRIBUTING.md](CONTRIBUTING.md) and the design is in [docs/architecture.md](docs/architecture.md).

Go is installed at `/home/rafael/go-dist/go/bin/go` (not on the system PATH by default):

```bash
export PATH="/home/rafael/go-dist/go/bin:$PATH"
```

- Keep the project general purpose: no subject matter from any particular use in code, comments, tests or docs.
- Docs are deliberately short. Put detail in `docs/`, not in the README.
- Commits end with the `Co-Authored-By` trailer given in the session context.
