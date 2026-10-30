# Arr-local fork notes

Private Arr experiments on `feat/arr-local` / `fork/arr-local` on top of upstream Ether.
Ship only to **origin** (`TheIndieArmy/ether`), not upstream PRs, until ready.

## Isolation

- Do **not** point Unraid Compose Manager **Ether** at this tree.
- Official runtime: `/mnt/user/appdata/ether` + `ghcr.io/braedonsaunders/ether`.
- Remotes: `upstream` = github.com/braedonsaunders/ether; `origin` = github.com/TheIndieArmy/ether; `local` = optional bare `../ether-private.git`.

## Goal

1. Owner settings: Sonarr/Radarr URL, API key, quality profile, root folder, path maps.
2. On Play: if Arr reports `hasFile`, stream the mapped local path (prefer over Usenet/debrid).
3. If missing: optional Arr search/ensure with progress, then fall back to Ether streaming.

## Status

### Done

- `config.Arr` with master **`enabled`** flag (unlocks connection fields when on).
- `internal/arr` client: status, roots, profiles, Radarr `MovieFileByTMDB`, Sonarr `EpisodeFileByShow`.
- Admin: `PUT /config/arr`, `POST /api/test/arr` (probe roots/profiles).
- Owner UI: **Integrations → Arr** — Enable → URL/key → Test → root/profile + path maps.
- **Play prefers Arr local:** `Discover` → path maps → `ProtocolLocal`; `WrapEngine` / `WrapIndexer` open/fetch from disk; resolver skips indexer-style gates for local releases.
- Local hits still win when a sticky Prefer points at a prior Usenet/debrid pick (Activity badge **Arr** when that path is used).
- Library helpers for missing-file Arr search (`Ensure`) exist; Play does **not** pass `EnsureMissing` yet.

### Next

- Opt-in / wire `EnsureMissing` on Play + progress UX when Arr is searching.
- Add-new-title flow (today `Ensure` only searches titles already in Arr).
- Rebuild embedded `internal/admin/dist` when shipping a binary (`cd web && npm run build` per upstream docs).

## Web HDR remux (fork test)

Upstream [#136](https://github.com/braedonsaunders/ether/pull/136) rejected browser-name HDR defaults. Branch `feat/web-hdr-evidence` (PR into `feat/arr-local`) promotes **Plex Web** HDR10 remux only when `X-Plex-Client-Profile-Extra` affirmatively lists `hevc` and does not deny 10-bit/PQ — aligned with observed PMS Firefox Direct Stream of HDR once HEVC decode is advertised.

To capture evidence for a future upstream PR: play Arr 4K HDR in Firefox with debug logging and save the decision line’s `profileExtra=` (sanitize tokens). A side-by-side PMS request/decision capture remains the bar Braedon asked for.
