# Copilot Instructions for manga-crawler-backend

## Project Overview
- This is a Go-based manga crawler and update notifier, focused on tracking manga updates from various sources and sending notifications via Telegram.
- Main entry point: [main.go](main.go). The app runs as a long-lived process, checking for updates every 15 minutes.
- Data is persisted in JSON files under [mangasdb/](mangasdb/), e.g., [readmanga.json](mangasdb/readmanga.json).

## Architecture & Key Components
- **Crawling/Parsing**: 
  - [parsers/](parsers/) contains site-specific parsers.
  - [sites/readmanga/](sites/readmanga/) implements crawling logic for the 'readmanga' and 'mint' sources. See [parser.go](sites/readmanga/parser.go) and [rss.go](sites/readmanga/rss.go).
- **Data Models**: 
  - [types/types.go](types/types.go) defines core types: `Manga`, `Chapter`, and RSS parsing structs.
- **Storage**: 
  - [storage/SaveMangaInfoToJson.go](storage/SaveMangaInfoToJson.go) handles reading/writing manga info to JSON, with mutex protection for concurrency.
- **Telegram Integration**: 
  - [telegram/telegram.go](telegram/telegram.go) manages bot initialization, message formatting, and notifications. Requires `.env` file with `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID`.
- **Utilities**: 
  - [utils/](utils/) provides helpers for user-agent rotation, chapter diffing, and RSS transformation.

## Developer Workflows
- **Build**: Standard Go build: `go build` in the project root.
- **Run**: Ensure `.env` is present, then run the binary or `go run main.go`.
- **Dependencies**: Managed via Go modules (`go.mod`).
- **Testing**: No explicit test files; manual testing via running the app and checking logs/Telegram.
- **Debugging**: Logs are printed to stdout and errors are sent to Telegram if enabled.

## Project-Specific Patterns & Conventions
- **Data Flow**: 
  - Crawler fetches RSS feeds, transforms them to `Manga` structs, diffs with stored data, and notifies via Telegram if new chapters are found.
- **Concurrency**: Storage uses a mutex to prevent race conditions during file writes.
- **Error Handling**: Most errors are logged and, for critical failures, also sent as Telegram messages.
- **Rate Limiting**: 4-second sleep between manga checks to avoid bans.
- **Localization**: Log and notification messages are in Russian.
- **Extending Sources**: To add new manga sources, update the lists in [sites/readmanga/parser.go](sites/readmanga/parser.go) or implement new parsers in [parsers/](parsers/).

## Integration Points
- **External**: Telegram Bot API, RSS feeds from manga sites.
- **Internal**: All cross-component communication is via Go function calls and shared types.

## Examples
- To add a new manga to track, append its slug to `readmangaList` or `mintList` in [sites/readmanga/parser.go](sites/readmanga/parser.go).
- To add a new site, implement a parser in [parsers/](parsers/) and wire it in [main.go](main.go).

---
For questions about project structure or extending functionality, see the referenced files for concrete patterns.
