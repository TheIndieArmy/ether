# Arr-local fork notes

Private `fork/arr-local` experiments on top of upstream Ether.

## Isolation

- Do **not** point Unraid Compose Manager **Ether** at this tree.
- Official runtime: `/mnt/user/appdata/ether` + `ghcr.io/braedonsaunders/ether`.
- Remotes: `upstream` = github.com/braedonsaunders/ether; `origin` = local `ether-private.git`.

## Goal

1. Owner settings: Sonarr/Radarr URL, API key, quality profile, root folder, path maps.
2. On Play: if Arr reports `hasFile`, stream the mapped local path (prefer over Usenet/debrid).
3. If missing: optional Arr search/ensure with progress, then fall back to Ether streaming.

## Status

Scaffold only — integration work starts next.
