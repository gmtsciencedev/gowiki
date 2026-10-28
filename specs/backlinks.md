# Backlinks

## Overview

The backlinks feature shows all pages that link to the current page. Two sources count as a link:

- an internal hyperlink (`[text](/path)`) to the page;
- for a row-bound page (`{database-row table=T}`), a `{database-query table=T}` directive on the other page: the query lists the table's rows and links to each row page when rendered.

This complements the existing include tracking (`IncludeIndex`).

## Backend

### Link extraction

`ExtractPageLinks(content, pagePath string) []string` in `backend/internal/markdown/links.go`:

- Reuses `linkRe` from `refs.go` (matches `[text](path)` but not `![alt](path)`)
- Filters to internal links only (skips `http://`, `https://`)
- Keeps only extension-less paths or `.md` paths (inverse of `addMediaRef` logic)
- Strips `.md` suffix, resolves via `ResolvePath`
- Skips content inside fenced code blocks
- Returns deduplicated, sorted list

### LinkIndex

`LinkIndex` in `backend/internal/storage/links.go`, following the `IncludeIndex` pattern:

- `PageToLinks map[string][]string` — maps source page to list of linked page paths
- Persisted as `data/meta/_links.json`
- `NewLinkIndex(basePath)`, `Load()`, `Save()`, `UpdatePage()`, `RemovePage()`
- `GetBacklinks(pagePath string) []string` — reverse lookup: returns all pages that link to the given page

### Wiring

- `FileStore` gains a `LinkIndex` field
- `Put()` extracts page links and calls `LinkIndex.UpdatePage()`
- `Delete()` calls `LinkIndex.RemovePage()`
- `RebuildIndexes()` rebuilds `LinkIndex` alongside other indexes
- `Put()`, `Move()` and `RebuildIndexes()` also record each page's queried tables (`ExtractDatabaseQueryTables`) with `LinkIndex.UpdatePageQueries()`. This map is held in memory only and rebuilt at startup; `_links.json` keeps its format.

### API

`GET /api/backlinks/{path}` — returns `{"backlinks": [{"path": "...", "title": "..."}]}`

- Registered in the optionalAuth + view permission group
- Handler looks up backlinks via `FileStore.GetBacklinks()` — hyperlink backlinks from `LinkIndex.GetBacklinks()`, plus, when the page carries `{database-row table=T}`, the pages from `LinkIndex.GetQueryPages(T)` — then reads each page to extract its title. Link rewriting on page moves uses the hyperlink backlinks only.

## Frontend

A "Backlinks" button is added to the view-mode action bar (after "History"). Clicking it:

1. Fetches `GET /api/backlinks/{pagePath}`
2. Enters history view (pushes browser state for back-button support)
3. Renders a simple list of linking pages with clickable links, showing both the page title and path (e.g. "My Page (/path/to/page)")

## Index persistence

The `_links.json` file is rebuilt from scratch on startup (like all other indexes) and updated incrementally on each page save/delete.
