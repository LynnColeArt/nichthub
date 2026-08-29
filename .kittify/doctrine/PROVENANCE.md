# Spec Kitty Doctrine — Fresh Project Seed

This `.kittify/doctrine/` tree was materialized by `spec-kitty charter
synthesize` running against a **fresh project** (no LLM-authored YAML under
`.kittify/charter/generated/`). It exists so `DoctrineService` discovers a
project layer and the runtime can advance; it is intentionally empty.

The runtime falls back to the packaged built-in doctrine
(`packs/built-in/`) for all artifact lookups until the LLM harness writes
project-local artifacts under `.kittify/charter/generated/` and you re-run
`spec-kitty charter synthesize`.

References
----------
- GitHub issue: https://github.com/Priivacy-ai/spec-kitty/issues/839
- Spec assumption A2: public CLI synthesize works on a fresh project.
- Project-root resolution: `src/charter/_doctrine_paths.py`.
