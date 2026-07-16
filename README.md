# Xtream Codes API

A Go library for the Xtream Codes API. It gives you a clean, type-safe interface over the API's sometimes inconsistent responses, and you can use it two ways:

- As a client: call the Xtream Codes endpoints and get back strongly-typed Go values.
- As a codec: import the dependency-light `types` package to marshal and unmarshal Xtream Codes payloads on their own.

## Installation

```bash
go get github.com/sherif-fanous/xtreamcodes
```

## Usage

### Initializing the Client

```go
import (
    "github.com/sherif-fanous/xtreamcodes"
)

// Basic client initialization
client := xtreamcodes.NewClient("https://your-iptv-server.com", "username", "password")

// With a custom HTTP client
httpClient := &http.Client{
    Timeout: 30 * time.Second,
}
client := xtreamcodes.NewClient(
    "https://your-iptv-server.com",
    "username",
    "password",
    xtreamcodes.WithHTTPClient(httpClient),
)
```

### Authentication Information

```go
authInfo, err := client.GetAuthInfo(context.Background())
if err != nil {
    // Handle error
}
fmt.Printf("User info: %+v\n", authInfo.UserInfo)
fmt.Printf("Server info: %+v\n", authInfo.ServerInfo)
```

### Live Streams

```go
// List all live categories
liveCategories, err := client.ListLiveCategories(context.Background())
if err != nil {
    // Handle error
}

// List all live streams
liveStreams, err := client.ListLiveStreams(context.Background())
if err != nil {
    // Handle error
}

// List live streams in a specific category
categoryID := 1  // Category ID
liveStreamsInCategory, err := client.ListLiveStreamsInCategory(context.Background(), categoryID)
if err != nil {
    // Handle error
}
```

### Series

```go
// List all series categories
seriesCategories, err := client.ListSeriesCategories(context.Background())
if err != nil {
    // Handle error
}

// List all series
seriesStreams, err := client.ListSeriesStreams(context.Background())
if err != nil {
    // Handle error
}

// Get detailed information about a specific series
seriesID := 456  // Series ID
seriesInfo, err := client.GetSeriesInfo(context.Background(), seriesID)
if err != nil {
    // Handle error
}
```

### Video on Demand (VOD)

```go
// List all VOD categories
vodCategories, err := client.ListVODCategories(context.Background())
if err != nil {
    // Handle error
}

// List all VOD streams
vodStreams, err := client.ListVODStreams(context.Background())
if err != nil {
    // Handle error
}

// Get detailed information about a specific VOD
vodID := 123  // VOD ID
vodInfo, err := client.GetVODInfo(context.Background(), vodID)
if err != nil {
    // Handle error
}
```

### EPG (Electronic Program Guide)

```go
// Get EPG data for a specific stream
streamID := 789  // Stream ID

// Get short EPG data
shortEPG, err := client.GetShortEPG(context.Background(), streamID)
if err != nil {
    // Handle error
}

// Get short EPG data with a limit
limit := 5  // Number of EPG entries to retrieve
shortEPGWithLimit, err := client.GetShortEPGWithLimits(context.Background(), streamID, limit)
if err != nil {
    // Handle error
}

// Get all EPG data for a stream
allEPG, err := client.GetAllEPG(context.Background(), streamID)
if err != nil {
    // Handle error
}

// Get XMLTV data
xmltv, err := client.GetXMLTV(context.Background())
if err != nil {
    // Handle error
}
```

### M3U Playlist

```go
// Get M3U playlist
playlist, err := client.GetPlaylist(
    context.Background(),
    xtreamcodes.PlaylistTypeM3UPlus,
    xtreamcodes.OutputFormatTS,
)
if err != nil {
    // Handle error
}
```

## Using the models without the client

```go
import "github.com/sherif-fanous/xtreamcodes/types"

// The types package depends only on internal/serde, so importing it avoids the
// HTTP client and its net/http, m3u, and xmltv dependencies.

// Decode a response into clean Go types.
var streams []types.LiveStream
if err := json.Unmarshal(body, &streams); err != nil {
    // Handle error
}

// Encode back into the Xtream Codes wire format.
out, err := json.Marshal(streams)
if err != nil {
    // Handle error
}

// xtreamcodes.LiveStream and types.LiveStream are the same type, so existing
// client code keeps working.
```
