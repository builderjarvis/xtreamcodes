# Changelog

## [0.2.1] - 2026-07-16

### Fixed

- Decode zero-date datetimes such as `"0000-00-00 00:00:00"` as a zero time instead of failing the entire response ([`f27767f`](https://github.com/sherif-fanous/xtreamcodes/commit/f27767f))
- Stop folding the server timezone into `ServerInfo.TimeNow` on decode, which could shift the timestamp on re-encode; `Timezone` remains available as its own field ([`f27767f`](https://github.com/sherif-fanous/xtreamcodes/commit/f27767f))

## [0.2.0] - 2026-07-14

_This release contains breaking API changes._

### Changed

- **Breaking:** require `PlaylistType` and `OutputFormat` values in `GetPlaylist` instead of strings ([`438b4ea`](https://github.com/sherif-fanous/xtreamcodes/commit/438b4ea))
- **Breaking:** require integer IDs in `GetAllEPG`, `GetShortEPG`, `GetShortEPGWithLimits`, `GetSeriesInfo`, `GetVODInfo`, and the `List*StreamsInCategory` methods instead of strings ([`438b4ea`](https://github.com/sherif-fanous/xtreamcodes/commit/438b4ea))
- Redact credentials from URLs stored in returned errors ([`438b4ea`](https://github.com/sherif-fanous/xtreamcodes/commit/438b4ea))

### Added

- Add `PlaylistType` and `OutputFormat` types with exported constants ([`438b4ea`](https://github.com/sherif-fanous/xtreamcodes/commit/438b4ea))

### Fixed

- Prevent a panic when decoding malformed duration strings ([`438b4ea`](https://github.com/sherif-fanous/xtreamcodes/commit/438b4ea))
- Return an error on invalid integer values instead of a `-1` sentinel ([`438b4ea`](https://github.com/sherif-fanous/xtreamcodes/commit/438b4ea))
- Marshal zero Unix timestamps as `null` instead of a negative number ([`438b4ea`](https://github.com/sherif-fanous/xtreamcodes/commit/438b4ea))
- Guard `Float64AsFloat` marshaling against a nil pointer ([`438b4ea`](https://github.com/sherif-fanous/xtreamcodes/commit/438b4ea))

## [0.0.1] - 2025-06-03

_First release._

[0.2.1]: https://github.com/sherif-fanous/xtreamcodes/releases/tag/v0.2.1
[0.2.0]: https://github.com/sherif-fanous/xtreamcodes/releases/tag/v0.2.0
[0.0.1]: https://github.com/sherif-fanous/xtreamcodes/releases/tag/v0.0.1
