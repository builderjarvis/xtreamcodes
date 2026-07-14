package xtreamcodes_test

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sherif-fanous/m3u"
	"github.com/sherif-fanous/xmltv"
	"github.com/sherif-fanous/xtreamcodes"
)

var (
	username string
	password string
)

func makePointer[T any](t T) *T {
	return &t
}

func makeURL(t *testing.T, s string) *url.URL {
	t.Helper()

	u, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	return u
}

func parseTime(t *testing.T, layout string, value string) time.Time {
	t.Helper()

	tt, err := time.Parse(layout, value)
	if err != nil {
		t.Fatal(err)
	}

	return tt
}

func TestMain(m *testing.M) {
	username = "user"
	password = "pass"

	os.Exit(m.Run())
}

func TestGetAuthInfo(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		data, err := os.ReadFile("testdata/get_info_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_info_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	authInfo, err := client.GetAuthInfo(context.Background())
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedAuthInfo := &xtreamcodes.AuthInfo{
		UserInfo: xtreamcodes.UserInfo{
			Username:             "user",
			Password:             "password",
			Message:              "Welcome to xtream-codes",
			IsAuthorized:         true,
			Status:               "Active",
			ExpiresAt:            time.Unix(1767225600, 0),
			IsTrial:              false,
			ActiveConnections:    0,
			CreatedAt:            time.Unix(1735689600, 0),
			MaxConnections:       1,
			AllowedOutputFormats: []string{"m3u8", "ts", "rtmp"},
		},
		ServerInfo: xtreamcodes.ServerInfo{
			URL:            "127.0.0.1",
			HTTPPort:       80,
			HTTPSPort:      443,
			ServerProtocol: "http",
			RTMPPort:       1935,
			Timezone:       "Etc/UTC",
			TimestampNow:   time.Unix(1735689600, 0),
			TimeNow:        parseTime(t, time.DateTime, "2025-01-01 00:00:00"),
			Process:        true,
		},
	}

	if diff := cmp.Diff(*expectedAuthInfo, *authInfo); diff != "" {
		t.Error(diff)
	}
}

func TestGetAllEPG(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("action") != "get_simple_data_table" {
			t.Errorf(
				"Expected action to be get_simple_data_table, got '%s'",
				r.URL.Query().Get("action"),
			)
		}

		if r.URL.Query().Get("stream_id") != "100" {
			t.Errorf("Expected stream_id to be 100, got '%s'", r.URL.Query().Get("stream_id"))
		}

		data, err := os.ReadFile("testdata/get_all_epg_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_all_epg_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	epg, err := client.GetAllEPG(context.Background(), 100)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedEPG := &xtreamcodes.EPG{
		EPGListings: []xtreamcodes.EPGListing{
			{
				ID:             1,
				EPGID:          1,
				Title:          "Title 1",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 00:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 01:00:00"),
				Description:    "Description 1",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735689600, 0),
				StopTimestamp:  time.Unix(1735693200, 0),
				StreamID:       makePointer(100),
			},
			{
				ID:             2,
				EPGID:          1,
				Title:          "Title 2",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 01:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 02:00:00"),
				Description:    "Description 2",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735693200, 0),
				StopTimestamp:  time.Unix(1735696800, 0),
				StreamID:       makePointer(100),
			},
			{
				ID:             3,
				EPGID:          1,
				Title:          "Title 3",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 02:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 03:00:00"),
				Description:    "Description 3",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735696800, 0),
				StopTimestamp:  time.Unix(1735700400, 0),
				StreamID:       makePointer(100),
			},
			{
				ID:             4,
				EPGID:          1,
				Title:          "Title 4",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 03:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 04:00:00"),
				Description:    "Description 4",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735700400, 0),
				StopTimestamp:  time.Unix(1735704000, 0),
				StreamID:       makePointer(100),
			},
		},
	}

	if diff := cmp.Diff(*expectedEPG, *epg); diff != "" {
		t.Error(diff)
	}
}

func TestGetPlaylist(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/get.php" {
			t.Errorf("Expected path to be /get.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("type") == "" {
			t.Error("Expected type to be set")
		}

		if r.URL.Query().Get("output") == "" {
			t.Error("Expected output to be set")
		}

		data, err := os.ReadFile("testdata/get_playlist_response.m3u8")
		if err != nil {
			t.Fatal("Failed to read testdata/get_playlist_response.m3u8:", err)
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	playlist, err := client.GetPlaylist(
		context.Background(),
		xtreamcodes.PlaylistTypeM3UPlus,
		xtreamcodes.OutputFormatTS,
	)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedPlaylist := &m3u.Playlist{
		Tracks: []m3u.Track{
			{
				Length:     -1,
				Name:       "Channel 1",
				TVGID:      makePointer("channel-1"),
				TVGName:    makePointer("Channel 1"),
				TVGLogo:    makeURL(t, "http://127.0.0.1/logos/live_stream_1.png"),
				GroupTitle: makePointer("Group 1"),
				URL:        makeURL(t, "http://127.0.0.1/user/pass/1"),
			},
			{
				Length:     -1,
				Name:       "Channel 2",
				TVGID:      makePointer("channel-2"),
				TVGName:    makePointer("Channel 2"),
				TVGLogo:    makeURL(t, "http://127.0.0.1/logos/live_stream_2.png"),
				GroupTitle: makePointer("Group 2"),
				URL:        makeURL(t, "http://127.0.0.1/user/pass/2"),
			},
		},
	}

	if diff := cmp.Diff(expectedPlaylist, playlist); diff != "" {
		t.Error(diff)
	}
}

func TestGetSeriesInfo(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("action") != "get_series_info" {
			t.Errorf("Expected action to be get_series_info, got '%s'", r.URL.Query().Get("action"))
		}

		if r.URL.Query().Get("series_id") != "1000" {
			t.Errorf("Expected series_id to be 1000, got '%s'", r.URL.Query().Get("series_id"))
		}

		data, err := os.ReadFile("testdata/get_series_info_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_series_info_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	seriesInfo, err := client.GetSeriesInfo(context.Background(), 1000)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedSeriesInfo := &xtreamcodes.Series{
		Episodes: map[string][]xtreamcodes.Episode{
			"1": {
				{
					AddedOn:            time.Unix(1735689600, 0),
					ContainerExtension: "mp4",
					CustomSID:          nil,
					DirectSource:       "",
					EpisodeInfo:        map[string]any{},
					EpisodeNumber:      1,
					ID:                 1,
					Season:             1,
					Title:              "Episode 1",
				},
			},
		},
		Info: xtreamcodes.SeriesInfo{
			BackdropPath:   []string{"http://127.0.0.1/logos/series_stream_1.png"},
			Cast:           "Actor 1, Actress 1, Actor 2",
			CategoryID:     makePointer(1),
			CategoryIDs:    []int{1},
			Cover:          "http://127.0.0.1/logos/series_stream_1.png",
			Director:       "Director 1, Director 2",
			EpisodeRunTime: 45,
			Genre:          "Drama",
			LastModifiedOn: time.Unix(1735689600, 0),
			Name:           "Series 1",
			Plot:           "Sunt commodo eu duis in veniam mollit deserunt ad aute. Sit ad duis id labore enim id minim. Deserunt sint laboris tempor esse nulla incididunt aliqua ipsum laboris est enim voluptate proident tempor ex. Adipisicing labore dolor velit adipisicing mollit aliqua amet velit sit. Id nulla id sunt ex labore.",
			Rating:         9,
			Rating5Based:   4.5,
			ReleaseDate:    parseTime(t, time.DateOnly, "2025-01-01"),
			TMDBID:         makePointer(1),
			YoutubeTrailer: "Z6WqiUh5QFwd",
		},
		Seasons: []xtreamcodes.Season{
			{
				AirDate:      parseTime(t, time.DateOnly, "2025-01-01"),
				Cover:        "http://127.0.0.1/logos/series_1.png",
				CoverBig:     "http://127.0.0.1/logos/series_1.png",
				CoverTMDB:    makePointer("http://127.0.0.1/logos/series_1.png"),
				Duration:     makePointer(45),
				EpisodeCount: 1,
				Name:         "Series 1",
				Overview:     "http://127.0.0.1/logos/series_1.png",
				ReleaseDate:  makePointer(parseTime(t, time.DateOnly, "2025-01-01")),
				SeasonNumber: 1,
			},
		},
	}

	if diff := cmp.Diff(*expectedSeriesInfo, *seriesInfo); diff != "" {
		t.Error(diff)
	}
}

func TestGetShortEPG(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("action") != "get_short_epg" {
			t.Errorf("Expected action to be get_short_epg, got '%s'", r.URL.Query().Get("action"))
		}

		if r.URL.Query().Get("stream_id") != "100" {
			t.Errorf("Expected stream_id to be 100, got '%s'", r.URL.Query().Get("stream_id"))
		}

		data, err := os.ReadFile("testdata/get_short_epg_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_short_epg_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	epg, err := client.GetShortEPG(context.Background(), 100)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedEPG := &xtreamcodes.EPG{
		EPGListings: []xtreamcodes.EPGListing{
			{
				ID:             1,
				EPGID:          1,
				Title:          "Title 1",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 00:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 01:00:00"),
				Description:    "Description 1",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735689600, 0),
				StopTimestamp:  time.Unix(1735693200, 0),
				StreamID:       makePointer(100),
			},
			{
				ID:             2,
				EPGID:          1,
				Title:          "Title 2",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 01:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 02:00:00"),
				Description:    "Description 2",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735693200, 0),
				StopTimestamp:  time.Unix(1735696800, 0),
				StreamID:       makePointer(100),
			},
			{
				ID:             3,
				EPGID:          1,
				Title:          "Title 3",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 02:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 03:00:00"),
				Description:    "Description 3",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735696800, 0),
				StopTimestamp:  time.Unix(1735700400, 0),
				StreamID:       makePointer(100),
			},
			{
				ID:             4,
				EPGID:          1,
				Title:          "Title 4",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 03:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 04:00:00"),
				Description:    "Description 4",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735700400, 0),
				StopTimestamp:  time.Unix(1735704000, 0),
				StreamID:       makePointer(100),
			},
		},
	}

	if diff := cmp.Diff(*expectedEPG, *epg); diff != "" {
		t.Error(diff)
	}
}

func TestGetShortEPGWithLimits(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("action") != "get_short_epg" {
			t.Errorf("Expected action to be get_short_epg, got '%s'", r.URL.Query().Get("action"))
		}

		if r.URL.Query().Get("stream_id") != "100" {
			t.Errorf("Expected stream_id to be 100, got '%s'", r.URL.Query().Get("stream_id"))
		}

		if r.URL.Query().Get("limit") != "3" {
			t.Errorf("Expected limit to be 3, got '%s'", r.URL.Query().Get("limit"))
		}

		data, err := os.ReadFile("testdata/get_short_epg_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_short_epg_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	epg, err := client.GetShortEPGWithLimits(context.Background(), 100, 3)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedEPG := &xtreamcodes.EPG{
		EPGListings: []xtreamcodes.EPGListing{
			{
				ID:             1,
				EPGID:          1,
				Title:          "Title 1",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 00:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 01:00:00"),
				Description:    "Description 1",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735689600, 0),
				StopTimestamp:  time.Unix(1735693200, 0),
				StreamID:       makePointer(100),
			},
			{
				ID:             2,
				EPGID:          1,
				Title:          "Title 2",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 01:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 02:00:00"),
				Description:    "Description 2",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735693200, 0),
				StopTimestamp:  time.Unix(1735696800, 0),
				StreamID:       makePointer(100),
			},
			{
				ID:             3,
				EPGID:          1,
				Title:          "Title 3",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 02:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 03:00:00"),
				Description:    "Description 3",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735696800, 0),
				StopTimestamp:  time.Unix(1735700400, 0),
				StreamID:       makePointer(100),
			},
			{
				ID:             4,
				EPGID:          1,
				Title:          "Title 4",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 03:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 04:00:00"),
				Description:    "Description 4",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735700400, 0),
				StopTimestamp:  time.Unix(1735704000, 0),
				StreamID:       makePointer(100),
			},
		},
	}

	if diff := cmp.Diff(*expectedEPG, *epg); diff != "" {
		t.Error(diff)
	}
}

func TestGetShortEPGWithLimitsInvalidLimits(t *testing.T) {
	t.Parallel()

	var serverCalled bool

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		serverCalled = true
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	_, err := client.GetShortEPGWithLimits(context.Background(), 100, 0)
	if err == nil {
		t.Fatal("Expected error for invalid limit, got nil")
	}

	if !strings.Contains(err.Error(), "limit must be greater than 0") {
		t.Errorf(
			"Expected error message to contain 'limit must be greater than 0', got '%s'",
			err.Error(),
		)
	}

	if serverCalled {
		t.Error(
			"Server was called despite invalid limit, client should validate before making request",
		)
	}
}

func TestGetVODInfo(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("action") != "get_vod_info" {
			t.Errorf("Expected action to be get_vod_info, got '%s'", r.URL.Query().Get("action"))
		}

		if r.URL.Query().Get("vod_id") != "1000" {
			t.Errorf("Expected vod_id to be 1000, got '%s'", r.URL.Query().Get("vod_id"))
		}

		data, err := os.ReadFile("testdata/get_vod_info_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_vod_info_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	vodInfo, err := client.GetVODInfo(context.Background(), 1000)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedVODInfo := &xtreamcodes.VOD{
		Info: xtreamcodes.VODInfo{
			Actors:       makePointer("Actor 1, Actress 1"),
			Age:          makePointer(""),
			Audio:        make(map[string]any),
			BackdropPath: []string{"http://127.0.0.1/logos/vod_stream_1.png"},
			Bitrate:      4096,
			Cast:         "Actor 1, Actress 1",
			Country:      makePointer(""),
			CoverBig:     makePointer("http://127.0.0.1/logos/vod_stream_1.png"),
			Description: makePointer(
				"Aliqua irure anim veniam officia minim deserunt dolor officia tempor deserunt. Anim magna aliquip sit enim voluptate Lorem duis cillum laboris commodo labore proident. Minim incididunt commodo laboris amet veniam quis. Labore sint velit nostrud nostrud Lorem labore amet nulla nostrud laborum tempor exercitation. Consectetur laborum Lorem officia ex consequat officia enim irure consectetur. Ea excepteur culpa ad ullamco anim magna eu nostrud fugiat id cupidatat.",
			),
			Director:        "Director 1",
			Duration:        time.Duration(9000 * time.Second),
			DurationSeconds: 9000,
			Genre:           "Action",
			MovieImage:      "http://127.0.0.1/logos/vod_stream_1.png",
			Name:            makePointer("Movie 1"),
			OriginalName:    makePointer("Movie 1"),
			Plot:            "Consequat reprehenderit quis tempor tempor dolore consequat consequat occaecat pariatur ad minim elit. Pariatur proident in fugiat deserunt fugiat tempor. Dolore mollit fugiat exercitation in. Voluptate irure amet enim sit voluptate do Lorem adipisicing consectetur cupidatat. Et ullamco ea occaecat incididunt non ut anim aliqua mollit pariatur Lorem dolore id.",
			Rating:          8.5,
			ReleaseDate:     parseTime(t, time.DateOnly, "2025-01-01"),
			Runtime:         makePointer(150),
			Status:          makePointer("Released"),
			TMDBID:          1,
			Video:           make(map[string]any),
		},
		MovieData: xtreamcodes.MovieData{
			AddedOn:            time.Unix(1735689600, 0),
			CategoryID:         makePointer(1),
			CategoryIDs:        []int{1},
			ContainerExtension: "mkv",
			CustomSID:          nil,
			DirectSource:       "",
			Name:               "Movie 1",
			StreamID:           1000,
		},
	}

	if diff := cmp.Diff(*expectedVODInfo, *vodInfo); diff != "" {
		t.Error(diff)
	}
}

func TestGetXMLTV(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/xmltv.php" {
			t.Errorf("Expected path to be /xmltv.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		data, err := os.ReadFile("testdata/get_xmltv_response.xml")
		if err != nil {
			t.Fatal("Failed to read testdata/get_xmltv_response.xml:", err)
		}

		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	epg, err := client.GetXMLTV(context.Background())
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedEPG := &xmltv.EPG{
		XMLName: xml.Name{
			Space: "",
			Local: "tv",
		},
		Date: &xmltv.Time{
			Time: parseTime(t, "20060102150405 -0700", "20220401000000 +0000"),
		},
		SourceInfoURL:     makePointer("example.com"),
		SourceInfoName:    makePointer("example"),
		SourceDataURL:     makePointer("example.com/a"),
		GeneratorInfoName: makePointer("Example Generator"),
		GeneratorInfoURL:  makePointer("https://example.com"),
		Channels: []xmltv.Channel{
			{
				XMLName: xml.Name{
					Space: "",
					Local: "channel",
				},
				ID: "channel-one.tv",
				DisplayNames: []xmltv.DisplayName{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "display-name",
						},
						Lang: makePointer("en"),
						Text: "Channel One",
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "display-name",
						},
						Lang: makePointer("fr"),
						Text: "Chaîne un",
					},
				},
				Icons: []xmltv.Icon{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "icon",
						},
						Source: "https://example.com/channel_one_icon.jpg",
						Width:  makePointer(100),
						Height: makePointer(100),
					},
				},
				URLs: []xmltv.URL{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "url",
						},
						System: makePointer("example"),
						Text:   "https://example.com/channel_one",
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "url",
						},
						System: makePointer("other_system"),
						Text:   "https://example.com/channel_one_alternate",
					},
				},
			},
			{
				XMLName: xml.Name{
					Space: "",
					Local: "channel",
				},
				ID: "channel-two.tv",
				DisplayNames: []xmltv.DisplayName{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "display-name",
						},
						Lang: nil,
						Text: "Channel Two: Minimum valid channel",
					},
				},
				Icons: nil,
				URLs:  nil,
			},
		},
		Programmes: []xmltv.Programme{
			{
				XMLName: xml.Name{
					Space: "",
					Local: "programme",
				},
				Start: xmltv.Time{
					Time: parseTime(t, "20060102150405 -0700", "20220331180000 +0000"),
				},
				Stop: &xmltv.Time{
					Time: parseTime(t, "20060102150405 -0700", "20220331190000 +0000"),
				},
				PDCStart: &xmltv.Time{
					Time: parseTime(t, "20060102150405 -0700", "20220331180000 +0000"),
				},
				VPSStart: &xmltv.Time{
					Time: parseTime(t, "20060102150405 -0700", "20220331180000 +0000"),
				},
				ShowView:   makePointer("12345"),
				VideoPlus:  makePointer("67890"),
				Channel:    "channel-one.tv",
				ClumpIndex: makePointer("0/1"),
				Titles: []xmltv.Title{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "title",
						},
						Lang: makePointer("en"),
						Text: "Programme One",
					},
				},
				SubTitles: []xmltv.SubTitle{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "sub-title",
						},
						Lang: makePointer("en"),
						Text: "Pilot",
					},
				},
				Descriptions: []xmltv.Description{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "desc",
						},
						Lang: makePointer("en"),
						Text: "This programme entry showcases all possible features of the DTD",
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "desc",
						},
						Lang: makePointer("en"),
						Text: "Short description",
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "desc",
						},
						Lang: makePointer("cy"),
						Text: "Mae'r cofnod rhaglen hwn yn arddangos holl nodweddion posibl y DTD",
					},
				},
				Credits: &xmltv.Credits{
					XMLName: xml.Name{
						Space: "",
						Local: "credits",
					},
					Directors: []xmltv.Director{
						{
							XMLName: xml.Name{
								Space: "",
								Local: "director",
							},
							Images: nil,
							URLs:   nil,
							Text:   "Samuel Jones",
						},
					},
					Actors: []xmltv.Actor{
						{
							XMLName: xml.Name{
								Space: "",
								Local: "actor",
							},
							Role:    makePointer("Walter Johnson"),
							IsGuest: nil,
							Images:  nil,
							URLs:    nil,
							Text:    "David Thompson",
						},
						{
							XMLName: xml.Name{
								Space: "",
								Local: "actor",
							},
							Role:    makePointer("Karl James"),
							IsGuest: makePointer(xmltv.Bool(true)),
							Images: []xmltv.Image{
								{
									XMLName: xml.Name{
										Space: "",
										Local: "image",
									},
									Type:        makePointer(xmltv.ImageTypePerson),
									Size:        nil,
									Orientation: nil,
									System:      nil,
									Text:        "\n                https://example.com/xxx.jpg",
								},
							},
							URLs: []xmltv.URL{
								{
									XMLName: xml.Name{
										Space: "",
										Local: "url",
									},
									System: makePointer("moviedb"),
									Text:   "\n                https://example.com/person/204",
								},
							},
							Text: " Ryan Lee \n        \n            ",
						},
					},
					Writers: []xmltv.Writer{
						{
							XMLName: xml.Name{
								Space: "",
								Local: "writer",
							},
							Images: nil,
							URLs:   nil,
							Text:   "Samuel Jones",
						},
					},
					Adapters: []xmltv.Adapter{
						{
							XMLName: xml.Name{
								Space: "",
								Local: "adapter",
							},
							Images: nil,
							URLs:   nil,
							Text:   "William Brown",
						},
					},
					Producers: []xmltv.Producer{
						{
							XMLName: xml.Name{
								Space: "",
								Local: "producer",
							},
							Images: nil,
							URLs:   nil,
							Text:   "Emily Davis",
						},
					},
					Composers: []xmltv.Composer{
						{
							XMLName: xml.Name{
								Space: "",
								Local: "composer",
							},
							Images: nil,
							URLs:   nil,
							Text:   "Max Wright",
						},
					},
					Editors: []xmltv.Editor{
						{
							XMLName: xml.Name{
								Space: "",
								Local: "editor",
							},
							Images: nil,
							URLs:   nil,
							Text:   "Nora Peterson",
						},
					},
					Presenters: []xmltv.Presenter{
						{
							XMLName: xml.Name{
								Space: "",
								Local: "presenter",
							},
							Images: nil,
							URLs:   nil,
							Text:   "Amanda Johnson",
						},
					},
					Commentators: []xmltv.Commentator{
						{
							XMLName: xml.Name{
								Space: "",
								Local: "commentator",
							},
							Images: nil,
							URLs:   nil,
							Text:   "James Wilson",
						},
					},
					Guests: []xmltv.Guest{
						{
							XMLName: xml.Name{
								Space: "",
								Local: "guest",
							},
							Images: nil,
							URLs:   nil,
							Text:   "Lucas Martin",
						},
						{
							XMLName: xml.Name{
								Space: "",
								Local: "guest",
							},
							Images: nil,
							URLs:   nil,
							Text:   "Emily Parker",
						},
						{
							XMLName: xml.Name{
								Space: "",
								Local: "guest",
							},
							Images: nil,
							URLs:   nil,
							Text:   "Oliver Nelson",
						},
					},
				},
				Date: &xmltv.Time{
					Time: parseTime(t, "20060102", "19901011"),
				},
				Categories: []xmltv.Category{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "category",
						},
						Lang: makePointer("en"),
						Text: "Comedy",
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "category",
						},
						Lang: makePointer("en"),
						Text: "Drama",
					},
				},
				Keywords: []xmltv.Keyword{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "keyword",
						},
						Lang: makePointer("en"),
						Text: "physical-comedy",
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "keyword",
						},
						Lang: makePointer("en"),
						Text: "romantic",
					},
				},
				Language: &xmltv.Language{
					XMLName: xml.Name{
						Space: "",
						Local: "language",
					},
					Lang: nil,
					Text: "English",
				},
				OriginalLanguage: &xmltv.OriginalLanguage{
					XMLName: xml.Name{
						Space: "",
						Local: "orig-language",
					},
					Lang: makePointer("en"),
					Text: "French",
				},
				Length: &xmltv.Length{
					XMLName: xml.Name{
						Space: "",
						Local: "length",
					},
					Units: xmltv.LengthUnitsMinutes,
					Text:  makePointer(60),
				},
				Icons: []xmltv.Icon{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "icon",
						},
						Source: "https://example.com/programme_one_icon.jpg",
						Width:  makePointer(100),
						Height: makePointer(100),
					},
				},
				URLs: []xmltv.URL{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "url",
						},
						System: makePointer("imdb"),
						Text:   "https://example.com/programme_one",
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "url",
						},
						System: nil,
						Text:   "https://example.com/programme_one_2",
					},
				},
				Countries: []xmltv.Country{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "country",
						},
						Lang: nil,
						Text: "US",
					},
				},
				EpisodeNumbers: []xmltv.EpisodeNumber{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "episode-num",
						},
						System: "onscreen",
						Text:   "S01E01",
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "episode-num",
						},
						System: "xmltv_ns",
						Text:   "1 . 1 . 0/1",
					},
				},
				Video: &xmltv.Video{
					XMLName: xml.Name{
						Space: "",
						Local: "video",
					},
					Present: makePointer(xmltv.Bool(true)),
					Colour:  makePointer(xmltv.Bool(false)),
					Aspect: &xmltv.Aspect{
						XMLName: xml.Name{
							Space: "",
							Local: "aspect",
						},
						Text: "16:9",
					},
					Quality: &xmltv.Quality{
						XMLName: xml.Name{
							Space: "",
							Local: "quality",
						},
						Text: "HDTV",
					},
				},
				Audio: &xmltv.Audio{
					XMLName: xml.Name{
						Space: "",
						Local: "audio",
					},
					Present: makePointer(xmltv.Bool(true)),
					Stereo: &xmltv.Stereo{
						XMLName: xml.Name{
							Space: "",
							Local: "stereo",
						},
						Text: "Dolby Digital",
					},
				},
				PreviouslyShown: &xmltv.PreviouslyShown{
					XMLName: xml.Name{
						Space: "",
						Local: "previously-shown",
					},
					Start: &xmltv.Time{
						Time: parseTime(t, "20060102150405 -0700", "20220331180000 +0000"),
					},
					Channel: makePointer("channel-two.tv"),
				},
				Premiere: &xmltv.Premiere{
					XMLName: xml.Name{
						Space: "",
						Local: "premiere",
					},
					Lang: nil,
					Text: "First time on British TV",
				},
				Lastchance: &xmltv.LastChance{
					XMLName: xml.Name{
						Space: "",
						Local: "last-chance",
					},
					Lang: makePointer("en"),
					Text: "Last time on this channel",
				},
				IsNew: xmltv.Bool(true),
				Subtitles: []xmltv.Subtitles{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "subtitles",
						},
						Type: makePointer(xmltv.SubtitlesTypeTeletext),
						Language: &xmltv.Language{
							XMLName: xml.Name{
								Space: "",
								Local: "language",
							},
							Lang: nil,
							Text: "English",
						},
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "subtitles",
						},
						Type: makePointer(xmltv.SubtitlesTypeOnScreen),
						Language: &xmltv.Language{
							XMLName: xml.Name{
								Space: "",
								Local: "language",
							},
							Lang: makePointer("en"),
							Text: "Spanish",
						},
					},
				},
				Ratings: []xmltv.Rating{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "rating",
						},
						System: makePointer("BBFC"),
						Value: &xmltv.Value{
							XMLName: xml.Name{
								Space: "",
								Local: "value",
							},
							Text: "15",
						},
						Icons: nil,
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "rating",
						},
						System: makePointer("MPAA"),
						Value: &xmltv.Value{
							XMLName: xml.Name{
								Space: "",
								Local: "value",
							},
							Text: "NC-17",
						},
						Icons: []xmltv.Icon{
							{
								XMLName: xml.Name{
									Space: "",
									Local: "icon",
								},
								Source: "NC-17_symbol.png",
								Width:  nil,
								Height: nil,
							},
						},
					},
				},
				StarRatings: []xmltv.StarRating{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "star-rating",
						},
						System: makePointer("TV Guide"),
						Value: &xmltv.Value{
							XMLName: xml.Name{
								Space: "",
								Local: "value",
							},
							Text: "4/5",
						},
						Icons: []xmltv.Icon{
							{
								XMLName: xml.Name{
									Space: "",
									Local: "icon",
								},
								Source: "stars.png",
								Width:  nil,
								Height: nil,
							},
						},
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "star-rating",
						},
						System: makePointer("IMDB"),
						Value: &xmltv.Value{
							XMLName: xml.Name{
								Space: "",
								Local: "value",
							},
							Text: "8/10",
						},
						Icons: nil,
					},
				},
				Reviews: []xmltv.Review{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "review",
						},
						Type:     makePointer(xmltv.ReviewTypeText),
						Source:   makePointer("Rotten Tomatoes"),
						Reviewer: makePointer("Joe Bloggs"),
						Lang:     makePointer("en"),
						Text:     "This is a\n            fantastic show!",
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "review",
						},
						Type:     makePointer(xmltv.ReviewTypeText),
						Source:   makePointer("IDMB"),
						Reviewer: makePointer("Jane Doe"),
						Lang:     makePointer("en"),
						Text:     "I love this show!",
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "review",
						},
						Type:     makePointer(xmltv.ReviewTypeURL),
						Source:   makePointer("Rotten Tomatoes"),
						Reviewer: makePointer("Joe Bloggs"),
						Lang:     makePointer("en"),
						Text:     "\n            https://example.com/programme_one_review",
					},
				},
				Images: []xmltv.Image{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "image",
						},
						Type:        makePointer(xmltv.ImageTypePoster),
						Size:        makePointer(xmltv.ImageSizeSmall),
						Orientation: makePointer(xmltv.ImageOrientationPortrait),
						System:      makePointer("tvdb"),
						Text:        "\n            https://tvdb.com/programme_one_poster_1.jpg",
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "image",
						},
						Type:        makePointer(xmltv.ImageTypePoster),
						Size:        makePointer(xmltv.ImageSizeMedium),
						Orientation: makePointer(xmltv.ImageOrientationPortrait),
						System:      makePointer("tmdb"),
						Text:        "\n            https://tmdb.com/programme_one_poster_2.jpg",
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "image",
						},
						Type:        makePointer(xmltv.ImageTypeBackdrop),
						Size:        makePointer(xmltv.ImageSizeLarge),
						Orientation: makePointer(xmltv.ImageOrientationLandscape),
						System:      makePointer("tvdb"),
						Text:        "\n            https://tvdb.com/programme_one_backdrop_3.jpg",
					},
					{
						XMLName: xml.Name{
							Space: "",
							Local: "image",
						},
						Type:        makePointer(xmltv.ImageTypeBackdrop),
						Size:        makePointer(xmltv.ImageSizeLarge),
						Orientation: makePointer(xmltv.ImageOrientationLandscape),
						System:      makePointer("tmdb"),
						Text:        "\n            https://tmdb.com/programme_one_backdrop_3.jpg",
					},
				},
			},
			{
				XMLName: xml.Name{
					Space: "",
					Local: "programme",
				},
				Start: xmltv.Time{
					Time: parseTime(t, "20060102150405 -0700", "20220331180000 +0000"),
				},
				Stop:       nil,
				PDCStart:   nil,
				VPSStart:   nil,
				ShowView:   nil,
				VideoPlus:  nil,
				Channel:    "channel-two.tv",
				ClumpIndex: nil,
				Titles: []xmltv.Title{
					{
						XMLName: xml.Name{
							Space: "",
							Local: "title",
						},
						Lang: nil,
						Text: "Programme Two: The minimum valid programme",
					},
				},
				SubTitles:        nil,
				Descriptions:     nil,
				Credits:          nil,
				Date:             nil,
				Categories:       nil,
				Keywords:         nil,
				Language:         nil,
				OriginalLanguage: nil,
				Length:           nil,
				Icons:            nil,
				URLs:             nil,
				Countries:        nil,
				EpisodeNumbers:   nil,
				Video:            nil,
				Audio:            nil,
				PreviouslyShown:  nil,
				Premiere:         nil,
				Lastchance:       nil,
				IsNew:            xmltv.Bool(false),
				Subtitles:        nil,
				Ratings:          nil,
				StarRatings:      nil,
				Reviews:          nil,
				Images:           nil,
			},
		},
	}

	if diff := cmp.Diff(*expectedEPG, *epg); diff != "" {
		t.Error(diff)
	}
}

func TestListLiveCategories(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("action") != "get_live_categories" {
			t.Errorf(
				"Expected action to be get_live_categories, got '%s'",
				r.URL.Query().Get("action"),
			)
		}

		data, err := os.ReadFile("testdata/get_live_categories_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_live_categories_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	liveCategories, err := client.ListLiveCategories(context.Background())
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedLiveCategories := []xtreamcodes.LiveCategory{
		{
			CategoryID:       1,
			CategoryName:     "Live Category 1",
			ParentCategoryID: 0,
		},
		{
			CategoryID:       2,
			CategoryName:     "Live Category 2",
			ParentCategoryID: 0,
		},
		{
			CategoryID:       3,
			CategoryName:     "Live Category 3",
			ParentCategoryID: 0,
		},
	}

	if diff := cmp.Diff(expectedLiveCategories, liveCategories); diff != "" {
		t.Error(diff)
	}
}

func TestListLiveStreams(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("action") != "get_live_streams" {
			t.Errorf(
				"Expected action to be get_live_streams, got '%s'",
				r.URL.Query().Get("action"),
			)
		}

		data, err := os.ReadFile("testdata/get_live_streams_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_live_streams_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	liveStreams, err := client.ListLiveStreams(context.Background())
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedLiveStreams := []xtreamcodes.LiveStream{
		{
			Number:              1,
			Name:                "Live Stream 1",
			StreamType:          "live",
			StreamID:            100,
			StreamIcon:          "http://127.0.0.1/logos/live_stream_1.png",
			EPGChannelID:        makePointer("LiveStream1"),
			AddedOn:             time.Unix(1735689600, 0),
			IsAdult:             false,
			CategoryID:          makePointer(1),
			CategoryIDs:         []int{1},
			CustomSID:           nil,
			HasCatchup:          false,
			DirectSource:        "",
			CatchupDurationDays: 0,
		},
		{
			Number:              2,
			Name:                "Live Stream 2",
			StreamType:          "live",
			StreamID:            101,
			StreamIcon:          "http://127.0.0.1/logos/live_stream_2.png",
			EPGChannelID:        makePointer("LiveStream2"),
			AddedOn:             time.Unix(1735689600, 0),
			IsAdult:             false,
			CategoryID:          makePointer(2),
			CategoryIDs:         []int{2},
			CustomSID:           nil,
			HasCatchup:          false,
			DirectSource:        "",
			CatchupDurationDays: 0,
		},
		{
			Number:              3,
			Name:                "Live Stream 3",
			StreamType:          "live",
			StreamID:            102,
			StreamIcon:          "http://127.0.0.1/logos/live_stream_3.png",
			EPGChannelID:        nil,
			AddedOn:             time.Unix(1735689600, 0),
			IsAdult:             false,
			CategoryID:          makePointer(3),
			CategoryIDs:         []int{3},
			CustomSID:           nil,
			HasCatchup:          false,
			DirectSource:        "",
			CatchupDurationDays: 0,
		},
	}

	if diff := cmp.Diff(expectedLiveStreams, liveStreams); diff != "" {
		t.Error(diff)
	}
}

func TestListLiveStreamsInCategory(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("action") != "get_live_streams" {
			t.Errorf(
				"Expected action to be get_live_streams, got '%s'",
				r.URL.Query().Get("action"),
			)
		}

		if r.URL.Query().Get("category_id") != "1" {
			t.Errorf("Expected category_id to be 1, got '%s'", r.URL.Query().Get("category_id"))
		}

		data, err := os.ReadFile("testdata/get_live_streams_in_category_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_live_streams_in_category_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	liveStreams, err := client.ListLiveStreamsInCategory(context.Background(), 1)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedLiveStreams := []xtreamcodes.LiveStream{
		{
			Number:              1,
			Name:                "Live Stream 1",
			StreamType:          "live",
			StreamID:            100,
			StreamIcon:          "http://127.0.0.1/logos/live_stream_1.png",
			EPGChannelID:        makePointer("LiveStream1"),
			AddedOn:             time.Unix(1735689600, 0),
			IsAdult:             false,
			CategoryID:          makePointer(1),
			CategoryIDs:         []int{1},
			CustomSID:           nil,
			HasCatchup:          false,
			DirectSource:        "",
			CatchupDurationDays: 0,
		},
		{
			Number:              2,
			Name:                "Live Stream 2",
			StreamType:          "live",
			StreamID:            101,
			StreamIcon:          "http://127.0.0.1/logos/live_stream_2.png",
			EPGChannelID:        makePointer("LiveStream2"),
			AddedOn:             time.Unix(1735689600, 0),
			IsAdult:             false,
			CategoryID:          makePointer(1),
			CategoryIDs:         []int{1},
			CustomSID:           nil,
			HasCatchup:          false,
			DirectSource:        "",
			CatchupDurationDays: 0,
		},
		{
			Number:              3,
			Name:                "Live Stream 3",
			StreamType:          "live",
			StreamID:            102,
			StreamIcon:          "http://127.0.0.1/logos/live_stream_3.png",
			EPGChannelID:        nil,
			AddedOn:             time.Unix(1735689600, 0),
			IsAdult:             false,
			CategoryID:          makePointer(1),
			CategoryIDs:         []int{1},
			CustomSID:           nil,
			HasCatchup:          false,
			DirectSource:        "",
			CatchupDurationDays: 0,
		},
	}

	if diff := cmp.Diff(expectedLiveStreams, liveStreams); diff != "" {
		t.Error(diff)
	}
}

func TestListSeriesCategories(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("action") != "get_series_categories" {
			t.Errorf(
				"Expected action to be get_series_categories, got '%s'",
				r.URL.Query().Get("action"),
			)
		}

		data, err := os.ReadFile("testdata/get_series_categories_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_series_categories_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	seriesCategories, err := client.ListSeriesCategories(context.Background())
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedSeriesCategories := []xtreamcodes.SeriesCategory{
		{
			CategoryID:       1,
			CategoryName:     "Series Category 1",
			ParentCategoryID: 0,
		},
		{
			CategoryID:       2,
			CategoryName:     "Series Category 2",
			ParentCategoryID: 0,
		},
		{
			CategoryID:       3,
			CategoryName:     "Series Category 3",
			ParentCategoryID: 0,
		},
	}

	if diff := cmp.Diff(expectedSeriesCategories, seriesCategories); diff != "" {
		t.Error(diff)
	}
}

func TestListSeriesStreams(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("action") != "get_series" {
			t.Errorf("Expected action to be get_series, got '%s'", r.URL.Query().Get("action"))
		}

		data, err := os.ReadFile("testdata/get_series_streams_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_series_streams_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	seriesStreams, err := client.ListSeriesStreams(context.Background())
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedSeriesStreams := []xtreamcodes.SeriesStream{
		{
			Number:         1,
			Name:           "Series Stream 1",
			SeriesID:       1000,
			Cover:          "http://127.0.0.1/logos/series_stream_1.png",
			Plot:           "Sunt commodo eu duis in veniam mollit deserunt ad aute. Sit ad duis id labore enim id minim. Deserunt sint laboris tempor esse nulla incididunt aliqua ipsum laboris est enim voluptate proident tempor ex. Adipisicing labore dolor velit adipisicing mollit aliqua amet velit sit. Id nulla id sunt ex labore.",
			Cast:           "Actor 1, Actress 1",
			Director:       "Director 1",
			Genre:          "Action",
			ReleaseDate:    parseTime(t, time.DateOnly, "2025-01-01"),
			LastModifiedOn: time.Unix(1735689600, 0),
			Rating:         10,
			Rating5Based:   5,
			BackdropPath:   []string{"http://127.0.0.1/logos/series_stream_1.jpg"},
			YoutubeTrailer: "Z6WqiUh5QFwd",
			TMDBID:         makePointer(1),
			EpisodeRunTime: 0,
			CategoryID:     makePointer(1),
			CategoryIDs:    []int{1},
		},
		{
			Number:         2,
			Name:           "Series Stream 2",
			SeriesID:       1001,
			Cover:          "http://127.0.0.1/logos/series_stream_2.png",
			Plot:           "Est velit minim commodo. In veniam amet eiusmod sint nostrud enim ea nisi consequat. Tempor nisi minim magna irure ad. Dolor ad anim duis in esse. Fugiat veniam ut id. Duis qui dolor sit amet in mollit. Esse ad tempor id pariatur laborum voluptate duis consectetur est Lorem.",
			Cast:           "Actor 2, Actress 2",
			Director:       "Director 2",
			Genre:          "Action",
			ReleaseDate:    parseTime(t, time.DateOnly, "2025-01-01"),
			LastModifiedOn: time.Unix(1735689600, 0),
			Rating:         8,
			Rating5Based:   4,
			BackdropPath:   []string{"http://127.0.0.1/logos/series_stream_2.jpg"},
			YoutubeTrailer: "aXgqPZelbHsK",
			TMDBID:         makePointer(2),
			EpisodeRunTime: 0,
			CategoryID:     makePointer(2),
			CategoryIDs:    []int{2},
		},
		{
			Number:         3,
			Name:           "Series Stream 3",
			SeriesID:       1002,
			Cover:          "http://127.0.0.1/logos/series_stream_3.png",
			Plot:           "Amet ea elit mollit enim dolore nisi amet veniam nisi. Et reprehenderit aute incididunt sint quis. Est officia non voluptate esse anim laboris officia elit sunt nulla proident elit. Ad do ea deserunt. Do est officia Lorem ut dolor minim eiusmod ipsum officia magna elit anim labore. Quis cupidatat ullamco do ipsum laborum nulla incididunt sit.",
			Cast:           "Actor 3, Actress 3",
			Director:       "Director 3",
			Genre:          "Action",
			ReleaseDate:    parseTime(t, time.DateOnly, "2025-01-01"),
			LastModifiedOn: time.Unix(1735689600, 0),
			Rating:         9,
			Rating5Based:   4.5,
			BackdropPath:   []string{"http://127.0.0.1/logos/series_stream_3.jpg"},
			YoutubeTrailer: "bzmfAawtFP37",
			TMDBID:         makePointer(3),
			EpisodeRunTime: 0,
			CategoryID:     makePointer(3),
			CategoryIDs:    []int{3},
		},
	}

	if diff := cmp.Diff(expectedSeriesStreams, seriesStreams); diff != "" {
		t.Error(diff)
	}
}

func TestListSeriesStreamsInCategory(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("action") != "get_series" {
			t.Errorf("Expected action to be get_series, got '%s'", r.URL.Query().Get("action"))
		}

		if r.URL.Query().Get("category_id") != "1" {
			t.Errorf("Expected category_id to be 1, got '%s'", r.URL.Query().Get("category_id"))
		}

		data, err := os.ReadFile("testdata/get_series_streams_in_category_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_series_streams_in_category_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	seriesStreams, err := client.ListSeriesStreamsInCategory(context.Background(), 1)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedSeriesStreams := []xtreamcodes.SeriesStream{
		{
			Number:         1,
			Name:           "Series Stream 1",
			SeriesID:       1000,
			Cover:          "http://127.0.0.1/logos/series_stream_1.png",
			Plot:           "Sunt commodo eu duis in veniam mollit deserunt ad aute. Sit ad duis id labore enim id minim. Deserunt sint laboris tempor esse nulla incididunt aliqua ipsum laboris est enim voluptate proident tempor ex. Adipisicing labore dolor velit adipisicing mollit aliqua amet velit sit. Id nulla id sunt ex labore.",
			Cast:           "Actor 1, Actress 1",
			Director:       "Director 1",
			Genre:          "Action",
			ReleaseDate:    parseTime(t, time.DateOnly, "2025-01-01"),
			LastModifiedOn: time.Unix(1735689600, 0),
			Rating:         10,
			Rating5Based:   5,
			BackdropPath:   []string{"http://127.0.0.1/logos/series_stream_1.jpg"},
			YoutubeTrailer: "Z6WqiUh5QFwd",
			TMDBID:         makePointer(1),
			EpisodeRunTime: 0,
			CategoryID:     makePointer(1),
			CategoryIDs:    []int{1},
		},
		{
			Number:         2,
			Name:           "Series Stream 2",
			SeriesID:       1001,
			Cover:          "http://127.0.0.1/logos/series_stream_2.png",
			Plot:           "Est velit minim commodo. In veniam amet eiusmod sint nostrud enim ea nisi consequat. Tempor nisi minim magna irure ad. Dolor ad anim duis in esse. Fugiat veniam ut id. Duis qui dolor sit amet in mollit. Esse ad tempor id pariatur laborum voluptate duis consectetur est Lorem.",
			Cast:           "Actor 2, Actress 2",
			Director:       "Director 2",
			Genre:          "Action",
			ReleaseDate:    parseTime(t, time.DateOnly, "2025-01-01"),
			LastModifiedOn: time.Unix(1735689600, 0),
			Rating:         8,
			Rating5Based:   4,
			BackdropPath:   []string{"http://127.0.0.1/logos/series_stream_2.jpg"},
			YoutubeTrailer: "aXgqPZelbHsK",
			TMDBID:         makePointer(2),
			EpisodeRunTime: 0,
			CategoryID:     makePointer(1),
			CategoryIDs:    []int{1},
		},
		{
			Number:         3,
			Name:           "Series Stream 3",
			SeriesID:       1002,
			Cover:          "http://127.0.0.1/logos/series_stream_3.png",
			Plot:           "Amet ea elit mollit enim dolore nisi amet veniam nisi. Et reprehenderit aute incididunt sint quis. Est officia non voluptate esse anim laboris officia elit sunt nulla proident elit. Ad do ea deserunt. Do est officia Lorem ut dolor minim eiusmod ipsum officia magna elit anim labore. Quis cupidatat ullamco do ipsum laborum nulla incididunt sit.",
			Cast:           "Actor 3, Actress 3",
			Director:       "Director 3",
			Genre:          "Action",
			ReleaseDate:    parseTime(t, time.DateOnly, "2025-01-01"),
			LastModifiedOn: time.Unix(1735689600, 0),
			Rating:         9,
			Rating5Based:   4.5,
			BackdropPath:   []string{"http://127.0.0.1/logos/series_stream_3.jpg"},
			YoutubeTrailer: "bzmfAawtFP37",
			TMDBID:         makePointer(3),
			EpisodeRunTime: 0,
			CategoryID:     makePointer(1),
			CategoryIDs:    []int{1},
		},
	}

	if diff := cmp.Diff(expectedSeriesStreams, seriesStreams); diff != "" {
		t.Error(diff)
	}
}

func TestListVODCategories(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("action") != "get_vod_categories" {
			t.Errorf(
				"Expected action to be get_vod_categories, got '%s'",
				r.URL.Query().Get("action"),
			)
		}

		data, err := os.ReadFile("testdata/get_vod_categories_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_vod_categories_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	vodCategories, err := client.ListVODCategories(context.Background())
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedVODCategories := []xtreamcodes.VODCategory{
		{
			CategoryID:       1,
			CategoryName:     "VOD Category 1",
			ParentCategoryID: 0,
		},
		{
			CategoryID:       2,
			CategoryName:     "VOD Category 2",
			ParentCategoryID: 0,
		},
		{
			CategoryID:       3,
			CategoryName:     "VOD Category 3",
			ParentCategoryID: 0,
		},
	}

	if diff := cmp.Diff(expectedVODCategories, vodCategories); diff != "" {
		t.Error(diff)
	}
}

func TestListVODStreams(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("action") != "get_vod_streams" {
			t.Errorf("Expected action to be get_vod_streams, got '%s'", r.URL.Query().Get("action"))
		}

		data, err := os.ReadFile("testdata/get_vod_streams_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_vod_streams_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	vodSstreams, err := client.ListVODStreams(context.Background())
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedVODStreams := []xtreamcodes.VODStream{
		{
			Number:             1,
			Name:               "VOD Stream 1",
			StreamType:         "movie",
			StreamID:           1000,
			StreamIcon:         "http://127.0.0.1/logos/vod_stream_1.png",
			Rating:             10,
			Rating5Based:       5,
			TMDBID:             makePointer(1),
			Trailer:            makePointer("earjOqLtrA1I"),
			AddedOn:            time.Unix(1735689600, 0),
			IsAdult:            false,
			CategoryID:         makePointer(1),
			CategoryIDs:        []int{1},
			ContainerExtension: "mp4",
			CustomSID:          nil,
			DirectSource:       "",
		},
		{
			Number:             2,
			Name:               "VOD Stream 2",
			StreamType:         "movie",
			StreamID:           1001,
			StreamIcon:         "http://127.0.0.1/logos/vod_stream_2.png",
			Rating:             8,
			Rating5Based:       4,
			TMDBID:             makePointer(2),
			Trailer:            makePointer("o9WQ1E1yO2Gz"),
			AddedOn:            time.Unix(1735689600, 0),
			IsAdult:            false,
			CategoryID:         makePointer(2),
			CategoryIDs:        []int{2},
			ContainerExtension: "mkv",
			CustomSID:          nil,
			DirectSource:       "",
		},
		{
			Number:             3,
			Name:               "VOD Stream 3",
			StreamType:         "movie",
			StreamID:           1002,
			StreamIcon:         "http://127.0.0.1/logos/vod_stream_3.png",
			Rating:             9,
			Rating5Based:       4.5,
			TMDBID:             makePointer(3),
			Trailer:            makePointer("u4AmVa6Zo5eQ"),
			AddedOn:            time.Unix(1735689600, 0),
			IsAdult:            false,
			CategoryID:         makePointer(3),
			CategoryIDs:        []int{3},
			ContainerExtension: "mp4",
			CustomSID:          nil,
			DirectSource:       "",
		},
	}

	if diff := cmp.Diff(expectedVODStreams, vodSstreams); diff != "" {
		t.Error(diff)
	}
}

func TestListVODStreamsInCategory(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/player_api.php" {
			t.Errorf("Expected path to be /player_api.php, got '%s'", r.URL.Path)
		}

		if r.URL.Query().Get("username") != username {
			t.Errorf(
				"Expected username to be %s, got '%s'",
				username,
				r.URL.Query().Get("username"),
			)
		}

		if r.URL.Query().Get("password") != password {
			t.Errorf(
				"Expected password to be %s, got '%s'",
				username,
				r.URL.Query().Get("password"),
			)
		}

		if r.URL.Query().Get("action") != "get_vod_streams" {
			t.Errorf("Expected action to be get_vod_streams, got '%s'", r.URL.Query().Get("action"))
		}

		if r.URL.Query().Get("category_id") != "1" {
			t.Errorf("Expected category_id to be 1, got '%s'", r.URL.Query().Get("category_id"))
		}

		data, err := os.ReadFile("testdata/get_vod_streams_in_category_response.json")
		if err != nil {
			t.Fatal("Failed to read testdata/get_vod_streams_in_category_response.json:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	vodStreams, err := client.ListVODStreamsInCategory(context.Background(), 1)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	expectedVODStreams := []xtreamcodes.VODStream{
		{
			Number:             1,
			Name:               "VOD Stream 1",
			StreamType:         "movie",
			StreamID:           1000,
			StreamIcon:         "http://127.0.0.1/logos/vod_stream_1.png",
			Rating:             10,
			Rating5Based:       5,
			TMDBID:             makePointer(1),
			Trailer:            makePointer("earjOqLtrA1I"),
			AddedOn:            time.Unix(1735689600, 0),
			IsAdult:            false,
			CategoryID:         makePointer(1),
			CategoryIDs:        []int{1},
			ContainerExtension: "mp4",
			CustomSID:          nil,
			DirectSource:       "",
		},
		{
			Number:             2,
			Name:               "VOD Stream 2",
			StreamType:         "movie",
			StreamID:           1001,
			StreamIcon:         "http://127.0.0.1/logos/vod_stream_2.png",
			Rating:             8,
			Rating5Based:       4,
			TMDBID:             makePointer(2),
			Trailer:            makePointer("o9WQ1E1yO2Gz"),
			AddedOn:            time.Unix(1735689600, 0),
			IsAdult:            false,
			CategoryID:         makePointer(1),
			CategoryIDs:        []int{1},
			ContainerExtension: "mkv",
			CustomSID:          nil,
			DirectSource:       "",
		},
		{
			Number:             3,
			Name:               "VOD Stream 3",
			StreamType:         "movie",
			StreamID:           1002,
			StreamIcon:         "http://127.0.0.1/logos/vod_stream_3.png",
			Rating:             9,
			Rating5Based:       4.5,
			TMDBID:             makePointer(3),
			Trailer:            makePointer("u4AmVa6Zo5eQ"),
			AddedOn:            time.Unix(1735689600, 0),
			IsAdult:            false,
			CategoryID:         makePointer(1),
			CategoryIDs:        []int{1},
			ContainerExtension: "mp4",
			CustomSID:          nil,
			DirectSource:       "",
		},
	}

	if diff := cmp.Diff(expectedVODStreams, vodStreams); diff != "" {
		t.Error(diff)
	}
}

func TestMarshalUnMarshalAuthInfo(t *testing.T) {
	t.Parallel()

	authInfo := &xtreamcodes.AuthInfo{
		UserInfo: xtreamcodes.UserInfo{
			Username:             "user",
			Password:             "password",
			Message:              "Welcome to xtream-codes",
			IsAuthorized:         true,
			Status:               "Active",
			ExpiresAt:            time.Unix(1767225600, 0),
			IsTrial:              false,
			ActiveConnections:    0,
			CreatedAt:            time.Unix(1735689600, 0),
			MaxConnections:       1,
			AllowedOutputFormats: []string{"m3u8", "ts", "rtmp"},
		},
		ServerInfo: xtreamcodes.ServerInfo{
			URL:            "127.0.0.1",
			HTTPPort:       80,
			HTTPSPort:      443,
			ServerProtocol: "http",
			RTMPPort:       1935,
			Timezone:       "Etc/UTC",
			TimestampNow:   time.Unix(1735689600, 0),
			TimeNow:        parseTime(t, time.DateTime, "2025-01-01 00:00:00"),
			Process:        true,
		},
	}

	data, err := json.Marshal(authInfo)
	if err != nil {
		t.Fatal("Failed to marshal AuthInfo:", err)
	}

	var unmarshaledAuthInfo xtreamcodes.AuthInfo
	if err := json.Unmarshal(data, &unmarshaledAuthInfo); err != nil {
		t.Fatal("Failed to unmarshal JSON back to AuthInfo:", err)
	}

	if diff := cmp.Diff(*authInfo, unmarshaledAuthInfo); diff != "" {
		t.Error("AuthInfo did not round-trip correctly:", diff)
	}
}

func TestMarshalUnMarshalCategories(t *testing.T) {
	t.Parallel()

	categories := []xtreamcodes.LiveCategory{
		{
			CategoryID:       1,
			CategoryName:     "Live Category 1",
			ParentCategoryID: 0,
		},
		{
			CategoryID:       2,
			CategoryName:     "Live Category 2",
			ParentCategoryID: 0,
		},
		{
			CategoryID:       3,
			CategoryName:     "Live Category 3",
			ParentCategoryID: 0,
		},
	}

	data, err := json.Marshal(categories)
	if err != nil {
		t.Fatal("Failed to marshal categories:", err)
	}

	var unmarshaledCategories []xtreamcodes.LiveCategory
	if err := json.Unmarshal(data, &unmarshaledCategories); err != nil {
		t.Fatal("Failed to unmarshal JSON back to []xtreamcodes.LiveCategory:", err)
	}

	if diff := cmp.Diff(categories, unmarshaledCategories); diff != "" {
		t.Error("categories did not round-trip correctly:", diff)
	}
}

func TestMarshalUnMarshalEPG(t *testing.T) {
	t.Parallel()

	epg := &xtreamcodes.EPG{
		EPGListings: []xtreamcodes.EPGListing{
			{
				ID:             1,
				EPGID:          1,
				Title:          "Title 1",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 00:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 01:00:00"),
				Description:    "Description 1",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735689600, 0),
				StopTimestamp:  time.Unix(1735693200, 0),
				StreamID:       makePointer(100),
			},
			{
				ID:             2,
				EPGID:          1,
				Title:          "Title 2",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 01:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 02:00:00"),
				Description:    "Description 2",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735693200, 0),
				StopTimestamp:  time.Unix(1735696800, 0),
				StreamID:       makePointer(100),
			},
			{
				ID:             3,
				EPGID:          1,
				Title:          "Title 3",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 02:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 03:00:00"),
				Description:    "Description 3",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735696800, 0),
				StopTimestamp:  time.Unix(1735700400, 0),
				StreamID:       makePointer(100),
			},
			{
				ID:             4,
				EPGID:          1,
				Title:          "Title 4",
				Language:       "",
				StartDateTime:  parseTime(t, time.DateTime, "2025-01-01 03:00:00"),
				EndDateTime:    parseTime(t, time.DateTime, "2025-01-01 04:00:00"),
				Description:    "Description 4",
				ChannelID:      "LiveStream1",
				StartTimestamp: time.Unix(1735700400, 0),
				StopTimestamp:  time.Unix(1735704000, 0),
				StreamID:       makePointer(100),
			},
		},
	}

	data, err := json.Marshal(epg)
	if err != nil {
		t.Fatal("Failed to marshal epg:", err)
	}

	var unmarshaledEPG xtreamcodes.EPG
	if err := json.Unmarshal(data, &unmarshaledEPG); err != nil {
		t.Fatal("Failed to unmarshal JSON back to xtreamcodes.EPG:", err)
	}

	if diff := cmp.Diff(*epg, unmarshaledEPG); diff != "" {
		t.Error("epg did not round-trip correctly:", diff)
	}
}

func TestMarshalUnMarshalLiveStreams(t *testing.T) {
	t.Parallel()

	liveStreams := []xtreamcodes.LiveStream{
		{
			Number:              1,
			Name:                "Live Stream 1",
			StreamType:          "live",
			StreamID:            100,
			StreamIcon:          "http://127.0.0.1/logos/live_stream_1.png",
			EPGChannelID:        makePointer("LiveStream1"),
			AddedOn:             time.Unix(1735689600, 0),
			IsAdult:             false,
			CategoryID:          makePointer(1),
			CategoryIDs:         []int{1},
			CustomSID:           nil,
			HasCatchup:          false,
			DirectSource:        "",
			CatchupDurationDays: 0,
		},
		{
			Number:              2,
			Name:                "Live Stream 2",
			StreamType:          "live",
			StreamID:            101,
			StreamIcon:          "http://127.0.0.1/logos/live_stream_2.png",
			EPGChannelID:        makePointer("LiveStream2"),
			AddedOn:             time.Unix(1735689600, 0),
			IsAdult:             false,
			CategoryID:          makePointer(2),
			CategoryIDs:         []int{2},
			CustomSID:           nil,
			HasCatchup:          false,
			DirectSource:        "",
			CatchupDurationDays: 0,
		},
		{
			Number:              3,
			Name:                "Live Stream 3",
			StreamType:          "live",
			StreamID:            102,
			StreamIcon:          "http://127.0.0.1/logos/live_stream_3.png",
			EPGChannelID:        nil,
			AddedOn:             time.Unix(1735689600, 0),
			IsAdult:             false,
			CategoryID:          makePointer(3),
			CategoryIDs:         []int{3},
			CustomSID:           nil,
			HasCatchup:          false,
			DirectSource:        "",
			CatchupDurationDays: 0,
		},
	}

	data, err := json.Marshal(liveStreams)
	if err != nil {
		t.Fatal("Failed to marshal liveStreams:", err)
	}

	var unmarshaledLiveStreams []xtreamcodes.LiveStream
	if err := json.Unmarshal(data, &unmarshaledLiveStreams); err != nil {
		t.Fatal("Failed to unmarshal JSON back to []xtreamcodes.LiveStream:", err)
	}

	if diff := cmp.Diff(liveStreams, unmarshaledLiveStreams); diff != "" {
		t.Error("liveStreams did not round-trip correctly:", diff)
	}
}

func TestMarshalUnmarshalSeriesInfo(t *testing.T) {
	t.Parallel()

	seriesInfo := &xtreamcodes.Series{
		Episodes: map[string][]xtreamcodes.Episode{
			"1": {
				{
					AddedOn:            time.Unix(1735689600, 0),
					ContainerExtension: "mp4",
					CustomSID:          nil,
					DirectSource:       "",
					EpisodeInfo:        map[string]any{},
					EpisodeNumber:      1,
					ID:                 1,
					Season:             1,
					Title:              "Episode 1",
				},
			},
		},
		Info: xtreamcodes.SeriesInfo{
			BackdropPath:   []string{"http://127.0.0.1/logos/series_stream_1.png"},
			Cast:           "Actor 1, Actress 1, Actor 2",
			CategoryID:     makePointer(1),
			CategoryIDs:    []int{1},
			Cover:          "http://127.0.0.1/logos/series_stream_1.png",
			Director:       "Director 1, Director 2",
			EpisodeRunTime: 45,
			Genre:          "Drama",
			LastModifiedOn: time.Unix(1735689600, 0),
			Name:           "Series 1",
			Plot:           "Sunt commodo eu duis in veniam mollit deserunt ad aute. Sit ad duis id labore enim id minim. Deserunt sint laboris tempor esse nulla incididunt aliqua ipsum laboris est enim voluptate proident tempor ex. Adipisicing labore dolor velit adipisicing mollit aliqua amet velit sit. Id nulla id sunt ex labore.",
			Rating:         9,
			Rating5Based:   4.5,
			ReleaseDate:    parseTime(t, time.DateOnly, "2025-01-01"),
			TMDBID:         makePointer(1),
			YoutubeTrailer: "Z6WqiUh5QFwd",
		},
		Seasons: []xtreamcodes.Season{
			{
				AirDate:      parseTime(t, time.DateOnly, "2025-01-01"),
				Cover:        "http://127.0.0.1/logos/series_1.png",
				CoverBig:     "http://127.0.0.1/logos/series_1.png",
				CoverTMDB:    makePointer("http://127.0.0.1/logos/series_1.png"),
				Duration:     makePointer(45),
				EpisodeCount: 1,
				Name:         "Series 1",
				Overview:     "http://127.0.0.1/logos/series_1.png",
				ReleaseDate:  makePointer(parseTime(t, time.DateOnly, "2025-01-01")),
				SeasonNumber: 1,
			},
		},
	}

	data, err := json.Marshal(seriesInfo)
	if err != nil {
		t.Fatal("Failed to marshal seriesInfo:", err)
	}

	var unmarshaledSeriesInfo xtreamcodes.Series
	if err := json.Unmarshal(data, &unmarshaledSeriesInfo); err != nil {
		t.Fatal("Failed to unmarshal JSON back to xtreamcodes.Series:", err)
	}

	if diff := cmp.Diff(*seriesInfo, unmarshaledSeriesInfo); diff != "" {
		t.Error("seriesInfo did not round-trip correctly:", diff)
	}
}

func TestMarshalUnMarshalSeriesStreams(t *testing.T) {
	t.Parallel()

	seriesStreams := []xtreamcodes.SeriesStream{
		{
			Number:         1,
			Name:           "Series Stream 1",
			SeriesID:       1000,
			Cover:          "http://127.0.0.1/logos/series_stream_1.png",
			Plot:           "Sunt commodo eu duis in veniam mollit deserunt ad aute. Sit ad duis id labore enim id minim. Deserunt sint laboris tempor esse nulla incididunt aliqua ipsum laboris est enim voluptate proident tempor ex. Adipisicing labore dolor velit adipisicing mollit aliqua amet velit sit. Id nulla id sunt ex labore.",
			Cast:           "Actor 1, Actress 1",
			Director:       "Director 1",
			Genre:          "Action",
			ReleaseDate:    parseTime(t, time.DateOnly, "2025-01-01"),
			LastModifiedOn: time.Unix(1735689600, 0),
			Rating:         10,
			Rating5Based:   5,
			BackdropPath:   []string{"http://127.0.0.1/logos/series_stream_1.jpg"},
			YoutubeTrailer: "Z6WqiUh5QFwd",
			TMDBID:         makePointer(1),
			EpisodeRunTime: 0,
			CategoryID:     makePointer(1),
			CategoryIDs:    []int{1},
		},
		{
			Number:         2,
			Name:           "Series Stream 2",
			SeriesID:       1001,
			Cover:          "http://127.0.0.1/logos/series_stream_2.png",
			Plot:           "Est velit minim commodo. In veniam amet eiusmod sint nostrud enim ea nisi consequat. Tempor nisi minim magna irure ad. Dolor ad anim duis in esse. Fugiat veniam ut id. Duis qui dolor sit amet in mollit. Esse ad tempor id pariatur laborum voluptate duis consectetur est Lorem.",
			Cast:           "Actor 2, Actress 2",
			Director:       "Director 2",
			Genre:          "Action",
			ReleaseDate:    parseTime(t, time.DateOnly, "2025-01-01"),
			LastModifiedOn: time.Unix(1735689600, 0),
			Rating:         8,
			Rating5Based:   4,
			BackdropPath:   []string{"http://127.0.0.1/logos/series_stream_2.jpg"},
			YoutubeTrailer: "aXgqPZelbHsK",
			TMDBID:         makePointer(2),
			EpisodeRunTime: 0,
			CategoryID:     makePointer(2),
			CategoryIDs:    []int{2},
		},
		{
			Number:         3,
			Name:           "Series Stream 3",
			SeriesID:       1002,
			Cover:          "http://127.0.0.1/logos/series_stream_3.png",
			Plot:           "Amet ea elit mollit enim dolore nisi amet veniam nisi. Et reprehenderit aute incididunt sint quis. Est officia non voluptate esse anim laboris officia elit sunt nulla proident elit. Ad do ea deserunt. Do est officia Lorem ut dolor minim eiusmod ipsum officia magna elit anim labore. Quis cupidatat ullamco do ipsum laborum nulla incididunt sit.",
			Cast:           "Actor 3, Actress 3",
			Director:       "Director 3",
			Genre:          "Action",
			ReleaseDate:    parseTime(t, time.DateOnly, "2025-01-01"),
			LastModifiedOn: time.Unix(1735689600, 0),
			Rating:         9,
			Rating5Based:   4.5,
			BackdropPath:   []string{"http://127.0.0.1/logos/series_stream_3.jpg"},
			YoutubeTrailer: "bzmfAawtFP37",
			TMDBID:         makePointer(3),
			EpisodeRunTime: 0,
			CategoryID:     makePointer(3),
			CategoryIDs:    []int{3},
		},
	}

	data, err := json.Marshal(seriesStreams)
	if err != nil {
		t.Fatal("Failed to marshal seriesStreams:", err)
	}

	var unmarshaledSeriesStreams []xtreamcodes.SeriesStream
	if err := json.Unmarshal(data, &unmarshaledSeriesStreams); err != nil {
		t.Fatal("Failed to unmarshal JSON back to []xtreamcodes.SeriesStream:", err)
	}

	if diff := cmp.Diff(seriesStreams, unmarshaledSeriesStreams); diff != "" {
		t.Error("seriesStreams did not round-trip correctly:", diff)
	}
}

func TestMarshalUnmarshalVODInfo(t *testing.T) {
	t.Parallel()

	vodInfo := &xtreamcodes.VOD{
		Info: xtreamcodes.VODInfo{
			Actors:       makePointer("Actor 1, Actress 1"),
			Age:          makePointer(""),
			Audio:        make(map[string]any),
			BackdropPath: []string{"http://127.0.0.1/logos/vod_stream_1.png"},
			Bitrate:      4096,
			Cast:         "Actor 1, Actress 1",
			Country:      makePointer(""),
			CoverBig:     makePointer("http://127.0.0.1/logos/vod_stream_1.png"),
			Description: makePointer(
				"Aliqua irure anim veniam officia minim deserunt dolor officia tempor deserunt. Anim magna aliquip sit enim voluptate Lorem duis cillum laboris commodo labore proident. Minim incididunt commodo laboris amet veniam quis. Labore sint velit nostrud nostrud Lorem labore amet nulla nostrud laborum tempor exercitation. Consectetur laborum Lorem officia ex consequat officia enim irure consectetur. Ea excepteur culpa ad ullamco anim magna eu nostrud fugiat id cupidatat.",
			),
			Director:        "Director 1",
			Duration:        time.Duration(9000 * time.Second),
			DurationSeconds: 9000,
			Genre:           "Action",
			MovieImage:      "http://127.0.0.1/logos/vod_stream_1.png",
			Name:            makePointer("Movie 1"),
			OriginalName:    makePointer("Movie 1"),
			Plot:            "Consequat reprehenderit quis tempor tempor dolore consequat consequat occaecat pariatur ad minim elit. Pariatur proident in fugiat deserunt fugiat tempor. Dolore mollit fugiat exercitation in. Voluptate irure amet enim sit voluptate do Lorem adipisicing consectetur cupidatat. Et ullamco ea occaecat incididunt non ut anim aliqua mollit pariatur Lorem dolore id.",
			Rating:          8.5,
			ReleaseDate:     parseTime(t, time.DateOnly, "2025-01-01"),
			Runtime:         makePointer(150),
			Status:          makePointer("Released"),
			TMDBID:          1,
			Video:           make(map[string]any),
		},
		MovieData: xtreamcodes.MovieData{
			AddedOn:            time.Unix(1735689600, 0),
			CategoryID:         makePointer(1),
			CategoryIDs:        []int{1},
			ContainerExtension: "mkv",
			CustomSID:          nil,
			DirectSource:       "",
			Name:               "Movie 1",
			StreamID:           1000,
		},
	}

	data, err := json.Marshal(vodInfo)
	if err != nil {
		t.Fatal("Failed to marshal vodInfo:", err)
	}

	var unmarshaledVODInfo xtreamcodes.VOD
	if err := json.Unmarshal(data, &unmarshaledVODInfo); err != nil {
		t.Fatal("Failed to unmarshal JSON back to xtreamcodes.VOD:", err)
	}

	if diff := cmp.Diff(*vodInfo, unmarshaledVODInfo); diff != "" {
		t.Error("vodInfo did not round-trip correctly:", diff)
	}
}

func TestMarshalUnMarshalVODStreams(t *testing.T) {
	t.Parallel()

	vodStreams := []xtreamcodes.VODStream{
		{
			Number:             1,
			Name:               "VOD Stream 1",
			StreamType:         "movie",
			StreamID:           1000,
			StreamIcon:         "http://127.0.0.1/logos/vod_stream_1.png",
			Rating:             10,
			Rating5Based:       5,
			TMDBID:             makePointer(1),
			Trailer:            makePointer("earjOqLtrA1I"),
			AddedOn:            time.Unix(1735689600, 0),
			IsAdult:            false,
			CategoryID:         makePointer(1),
			CategoryIDs:        []int{1},
			ContainerExtension: "mp4",
			CustomSID:          nil,
			DirectSource:       "",
		},
		{
			Number:             2,
			Name:               "VOD Stream 2",
			StreamType:         "movie",
			StreamID:           1001,
			StreamIcon:         "http://127.0.0.1/logos/vod_stream_2.png",
			Rating:             8,
			Rating5Based:       4,
			TMDBID:             makePointer(2),
			Trailer:            makePointer("o9WQ1E1yO2Gz"),
			AddedOn:            time.Unix(1735689600, 0),
			IsAdult:            false,
			CategoryID:         makePointer(2),
			CategoryIDs:        []int{2},
			ContainerExtension: "mkv",
			CustomSID:          nil,
			DirectSource:       "",
		},
		{
			Number:             3,
			Name:               "VOD Stream 3",
			StreamType:         "movie",
			StreamID:           1002,
			StreamIcon:         "http://127.0.0.1/logos/vod_stream_3.png",
			Rating:             9,
			Rating5Based:       4.5,
			TMDBID:             makePointer(3),
			Trailer:            makePointer("u4AmVa6Zo5eQ"),
			AddedOn:            time.Unix(1735689600, 0),
			IsAdult:            false,
			CategoryID:         makePointer(3),
			CategoryIDs:        []int{3},
			ContainerExtension: "mp4",
			CustomSID:          nil,
			DirectSource:       "",
		},
	}

	data, err := json.Marshal(vodStreams)
	if err != nil {
		t.Fatal("Failed to marshal vodStreams:", err)
	}

	var unmarshaledVODStreams []xtreamcodes.VODStream
	if err := json.Unmarshal(data, &unmarshaledVODStreams); err != nil {
		t.Fatal("Failed to unmarshal JSON back to []xtreamcodes.VODStream:", err)
	}

	if diff := cmp.Diff(vodStreams, unmarshaledVODStreams); diff != "" {
		t.Error("vodStreams did not round-trip correctly:", diff)
	}
}

func TestRequestHTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write(
			[]byte(
				`<!doctype html><html lang="en"><head><title>503 Service Unavailable</title></head><body><h1>503 Service Unavailable</h1><p>The server was unable to complete your request. Please try again later.</p></body></html>`,
			),
		)
	}))
	defer server.Close()

	client := xtreamcodes.NewClient(server.URL, username, password)
	_, err := client.GetAuthInfo(context.Background())
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	var httpError *xtreamcodes.HTTPError
	if !errors.As(err, &httpError) {
		t.Fatalf("Expected error to be of type *xtreamcodes.HTTPError, got %T", err)
	}

	// Test the Error() method
	errorMsg := httpError.Error()

	if !strings.Contains(errorMsg, strconv.Itoa(http.StatusServiceUnavailable)) {
		t.Errorf("Expected error message to contain status code %d, got: %d",
			http.StatusServiceUnavailable, httpError.StatusCode)
	}

	if !strings.Contains(errorMsg, "503 Service Unavailable") {
		t.Errorf("Expected error message to contain '503 Service Unavailable', got: %s", errorMsg)
	}

	if !strings.Contains(errorMsg, server.URL) {
		t.Errorf(
			"Expected error message to contain the request URL %s got: %s",
			server.URL,
			httpError.URL,
		)
	}
}
