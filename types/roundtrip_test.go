package types_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/sherif-fanous/xtreamcodes/types"
)

// roundTrip asserts that a fixture survives the codec without loss, which is the
// contract a proxy (acting as an Xtream Codes server) relies on:
//
//   - model fidelity: decode -> encode -> decode yields an equal model, so no
//     field is dropped or corrupted on the way back out.
//   - wire idempotence: encoding the re-decoded model reproduces the exact same
//     bytes, so the emitted representation is stable.
func roundTrip[T any](t *testing.T, fixture string) {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "testdata", fixture))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var m1 T
	if err := json.Unmarshal(raw, &m1); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	b1, err := json.Marshal(&m1)
	if err != nil {
		t.Fatalf("encode model: %v", err)
	}

	var m2 T
	if err := json.Unmarshal(b1, &m2); err != nil {
		t.Fatalf("decode re-encoded model: %v", err)
	}

	if !reflect.DeepEqual(m1, m2) {
		t.Errorf("model changed across round-trip\n first: %+v\nsecond: %+v", m1, m2)
	}

	b2, err := json.Marshal(&m2)
	if err != nil {
		t.Fatalf("re-encode model: %v", err)
	}

	if !bytes.Equal(b1, b2) {
		t.Errorf("wire representation not idempotent\nfirst:  %s\nsecond: %s", b1, b2)
	}
}

func TestRoundTrip(t *testing.T) {
	t.Run("AuthInfo", func(t *testing.T) { roundTrip[types.AuthInfo](t, "get_info_response.json") })

	t.Run("AllEPG", func(t *testing.T) { roundTrip[types.EPG](t, "get_all_epg_response.json") })
	t.Run("ShortEPG", func(t *testing.T) { roundTrip[types.EPG](t, "get_short_epg_response.json") })

	t.Run("LiveCategories", func(t *testing.T) {
		roundTrip[[]types.Category](t, "get_live_categories_response.json")
	})
	t.Run("SeriesCategories", func(t *testing.T) {
		roundTrip[[]types.Category](t, "get_series_categories_response.json")
	})
	t.Run("VODCategories", func(t *testing.T) {
		roundTrip[[]types.Category](t, "get_vod_categories_response.json")
	})

	t.Run("LiveStreams", func(t *testing.T) {
		roundTrip[[]types.LiveStream](t, "get_live_streams_response.json")
	})
	t.Run("LiveStreamsInCategory", func(t *testing.T) {
		roundTrip[[]types.LiveStream](t, "get_live_streams_in_category_response.json")
	})

	t.Run("SeriesStreams", func(t *testing.T) {
		roundTrip[[]types.SeriesStream](t, "get_series_streams_response.json")
	})
	t.Run("SeriesStreamsInCategory", func(t *testing.T) {
		roundTrip[[]types.SeriesStream](t, "get_series_streams_in_category_response.json")
	})
	t.Run("SeriesInfo", func(t *testing.T) {
		roundTrip[types.Series](t, "get_series_info_response.json")
	})

	t.Run("VODStreams", func(t *testing.T) {
		roundTrip[[]types.VODStream](t, "get_vod_streams_response.json")
	})
	t.Run("VODStreamsInCategory", func(t *testing.T) {
		roundTrip[[]types.VODStream](t, "get_vod_streams_in_category_response.json")
	})
	t.Run("VODInfo", func(t *testing.T) { roundTrip[types.VOD](t, "get_vod_info_response.json") })
}
