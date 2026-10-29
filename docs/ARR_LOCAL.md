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

### Done (settings slice)

- `config.Arr` with master **`enabled`** flag (unlocks connection fields when on).
- `internal/arr` client: status, root folders, quality profiles, Radarr `MovieFileByTMDB`.
- Admin: `PUT /config/arr`, `POST /api/test/arr` (returns probe with roots/profiles).
- Owner UI: **Integrations → Arr** — Enable Arrs → Sonarr/Radarr URL/key → Test unlocks root/profile picks + path maps.

### Next

- Wire playback to prefer Arr `hasFile` after path-map rewrite.
- Missing-title Arr search/ensure + progress UX.
- Rebuild embedded `internal/admin/dist` when shipping a binary (`cd web && npm run build` per upstream docs).
