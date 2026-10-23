package xtreamcodes

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sherif-fanous/m3u"
	"github.com/sherif-fanous/xmltv"
)

const (
	defaultHTTPClientTimeout = 30 * time.Second
	defaultUserAgent         = "go-xtreamcodes"
	defaultMaxResponseBytes  = 64 << 20

	apiPath      = "/player_api.php"
	epgPath      = "/xmltv.php"
	playlistPath = "/get.php"

	allEPGAction           = "get_simple_data_table"
	liveCategoriesAction   = "get_live_categories"
	liveStreamsAction      = "get_live_streams"
	seriesCategoriesAction = "get_series_categories"
	seriesInfoAction       = "get_series_info"
	seriesStreamsAction    = "get_series"
	shortEPGAction         = "get_short_epg"
	vodCategoriesAction    = "get_vod_categories"
	vodInfoAction          = "get_vod_info"
	vodStreamsAction       = "get_vod_streams"
)

// PlaylistType represents the type of playlist requested from the Xtream Codes API.
type PlaylistType string

const (
	PlaylistTypeM3U     PlaylistType = "m3u"
	PlaylistTypeM3UPlus PlaylistType = "m3u_plus"
)

// OutputFormat represents the container/output format requested for a playlist.
type OutputFormat string

const (
	OutputFormatTS   OutputFormat = "ts"
	OutputFormatM3U8 OutputFormat = "m3u8"
	OutputFormatMP4  OutputFormat = "mp4"
)

// Client represents a client for the Xtream Codes API.
type Client struct {
	host             string
	username         string
	password         string
	httpClient       *http.Client
	userAgent        string
	maxResponseBytes int64
}

// Option represents a function that configures the Client using the functional options pattern.
type Option func(*Client)

// NewClient creates a new Client instance with the specified host, username, password, and options.
// The Client will use the default HTTP client with a timeout of 30 seconds.
//
// To set the user agent, use the WithUserAgent option.
// To customize the HTTP client, use the WithHTTPClient option.
func NewClient(host string, username string, password string, options ...Option) *Client {
	client := &Client{
		host:     host,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: defaultHTTPClientTimeout,
		},
		userAgent:        defaultUserAgent,
		maxResponseBytes: defaultMaxResponseBytes,
	}

	for _, option := range options {
		if option != nil {
			option(client)
		}
	}

	return client
}

// WithHTTPClient sets the HTTP client for the client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// WithMaxResponseBytes sets the maximum decoded response size in bytes.
// Nonpositive values retain the default limit of 64 MiB.
func WithMaxResponseBytes(limit int64) Option {
	return func(c *Client) {
		if limit > 0 {
			c.maxResponseBytes = limit
		}
	}
}

// WithUserAgent sets the user agent for the client.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// decodeResponse executes the request and decodes the JSON response body into a
// value of type T. It is the single place the request/decode/error-wrapping
// logic lives so the various Get/List methods don't each repeat it.
func decodeResponse[T any](ctx context.Context, client *Client, reqURL string) (*T, error) {
	resp, err := client.doRequest(ctx, reqURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var v T
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&v); err != nil {
		return nil, &DecoderError{
			URL:                    redactURL(reqURL),
			UnderlyingDecoderError: safeCause(err),
		}
	}
	// Reject trailing values and read to EOF so a response cannot evade the size limit.
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			err = errors.New("multiple JSON values")
		}
		return nil, &DecoderError{URL: redactURL(reqURL), UnderlyingDecoderError: safeCause(err)}
	}

	return &v, nil
}

// listStreams is implemented as a standalone generic function rather than a method
// because Go doesn't support generic methods on receiver types. It's a thin
// wrapper over decodeResponse so the List*Streams call sites stay a single line.
func listStreams[E LiveStream | SeriesStream | VODStream](
	ctx context.Context,
	client *Client,
	reqURL string,
) ([]E, error) {
	streams, err := decodeResponse[[]E](ctx, client, reqURL)
	if err != nil {
		return nil, err
	}

	return *streams, nil
}

// redactURL masks the username and password query parameters in a URL so that
// credentials are never stored in errors or leaked into logs.
func redactURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "[invalid URL]"
	}
	parsed.User = nil
	parsed.Fragment = ""

	query := parsed.Query()
	if query.Has("username") {
		query.Set("username", "REDACTED")
	}
	if query.Has("password") {
		query.Set("password", "REDACTED")
	}
	parsed.RawQuery = query.Encode()

	return parsed.String()
}

// buildURL constructs the full URL for the API request, including the host, path, and query parameters.
func (c *Client) buildURL(path string, queryParameters url.Values) string {
	base, err := url.Parse(c.host)
	if err != nil || base.Hostname() == "" || (base.Scheme != "http" && base.Scheme != "https") || base.User != nil || base.RawQuery != "" || base.ForceQuery || base.Fragment != "" || base.Opaque != "" {
		return ""
	}
	base.Path = strings.TrimRight(base.Path, "/") + path
	base.RawPath = ""
	credentialsQueryParameters := url.Values{}
	credentialsQueryParameters.Add("username", c.username)
	credentialsQueryParameters.Add("password", c.password)

	for key, values := range queryParameters {
		if key != "username" && key != "password" {
			credentialsQueryParameters[key] = values
		}
	}
	base.RawQuery = credentialsQueryParameters.Encode()

	return base.String()
}

// doRequest executes an HTTP GET request to the specified URL and returns the response.
func (c *Client) doRequest(ctx context.Context, reqURL string) (*http.Response, error) {
	if reqURL == "" {
		return nil, errors.New("invalid host: require an HTTP(S) URL without userinfo, query or fragment")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", safeCause(err))
	}

	req.Header.Set("User-Agent", c.userAgent)

	// Copy the client to avoid changing the caller's redirect policy or racing
	// concurrent requests. Never send credential-bearing redirects to another origin.
	httpClient := *c.httpClient
	checkRedirect := httpClient.CheckRedirect
	httpClient.CheckRedirect = func(next *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("too many redirects")
		}
		if next.URL.Scheme != req.URL.Scheme || !strings.EqualFold(next.URL.Host, req.URL.Host) || next.URL.User != nil {
			return errors.New("cross-origin redirect blocked")
		}
		// The automatic Referer can contain credentials from the previous URL.
		next.Header.Del("Referer")
		if checkRedirect != nil {
			if err := checkRedirect(next, via); err != nil {
				return err
			}
			if next.URL.Scheme != req.URL.Scheme || !strings.EqualFold(next.URL.Host, req.URL.Host) || next.URL.User != nil {
				return errors.New("cross-origin redirect blocked")
			}
			next.Header.Del("Referer")
		}
		return nil
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", safeCause(err))
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()

		return nil, &HTTPError{
			URL:        redactURL(reqURL),
			StatusCode: resp.StatusCode,
		}
	}
	if resp.ContentLength > c.maxResponseBytes {
		_ = resp.Body.Close()
		return nil, ErrResponseTooLarge
	}
	resp.Body = &boundedBody{ReadCloser: resp.Body, remaining: c.maxResponseBytes, ctx: ctx}

	return resp, nil
}

// getEPG retrieves EPG data based on the provided query parameters.
// It is a helper function used by GetAllEPG, GetShortEPG, and GetShortEPGWithLimits methods
// to avoid code duplication.
func (c *Client) getEPG(ctx context.Context, queryParameters url.Values) (*EPG, error) {
	return decodeResponse[EPG](ctx, c, c.buildURL(apiPath, queryParameters))
}

// listCategories retrieves a list of categories from the specified URL.
// It is a helper function used by ListLiveCategories, ListSeriesCategories, and ListVODCategories
// methods to avoid code duplication.
func (c *Client) listCategories(ctx context.Context, reqURL string) ([]Category, error) {
	categories, err := decodeResponse[[]Category](ctx, c, reqURL)
	if err != nil {
		return nil, err
	}

	return *categories, nil
}

// GetAllEPG retrieves all EPG data for a specific stream ID.
func (c *Client) GetAllEPG(ctx context.Context, streamID int) (*EPG, error) {
	queryParameters := url.Values{}
	queryParameters.Add("action", allEPGAction)
	queryParameters.Add("stream_id", strconv.Itoa(streamID))

	return c.getEPG(ctx, queryParameters)
}

// GetAuthInfo retrieves authentication information.
func (c *Client) GetAuthInfo(ctx context.Context) (*AuthInfo, error) {
	return decodeResponse[AuthInfo](ctx, c, c.buildURL(apiPath, nil))
}

// GetPlaylist retrieves a M3U/M3U8 playlist.
func (c *Client) GetPlaylist(
	ctx context.Context,
	playlistType PlaylistType,
	outputFormat OutputFormat,
) (*m3u.Playlist, error) {
	queryParameters := url.Values{}
	queryParameters.Add("type", string(playlistType))
	queryParameters.Add("output", string(outputFormat))

	reqURL := c.buildURL(playlistPath, queryParameters)

	resp, err := c.doRequest(ctx, reqURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var playlist m3u.Playlist
	if err := m3u.NewDecoder(resp.Body).Decode(&playlist); err != nil {
		return nil, &DecoderError{
			URL:                    redactURL(reqURL),
			UnderlyingDecoderError: safeCause(err),
		}
	}
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return nil, safeCause(err)
	}

	return &playlist, nil
}

// GetShortEPG retrieves a short EPG for a specific stream ID.
func (c *Client) GetShortEPG(ctx context.Context, streamID int) (*EPG, error) {
	queryParameters := url.Values{}
	queryParameters.Add("action", shortEPGAction)
	queryParameters.Add("stream_id", strconv.Itoa(streamID))

	return c.getEPG(ctx, queryParameters)
}

// GetShortEPGWithLimits retrieves a short EPG for a specific stream ID with a limit on the number of entries.
func (c *Client) GetShortEPGWithLimits(
	ctx context.Context,
	streamID int,
	limit int,
) (*EPG, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be greater than 0")
	}

	queryParameters := url.Values{}
	queryParameters.Add("action", shortEPGAction)
	queryParameters.Add("stream_id", strconv.Itoa(streamID))
	queryParameters.Add("limit", strconv.Itoa(limit))

	return c.getEPG(ctx, queryParameters)
}

// GetSeriesInfo retrieves information about a specific series by its ID.
func (c *Client) GetSeriesInfo(ctx context.Context, seriesID int) (*Series, error) {
	queryParameters := url.Values{}
	queryParameters.Add("action", seriesInfoAction)
	queryParameters.Add("series_id", strconv.Itoa(seriesID))

	return decodeResponse[Series](ctx, c, c.buildURL(apiPath, queryParameters))
}

// GetVODInfo retrieves information about a specific VOD by its ID.
func (c *Client) GetVODInfo(ctx context.Context, vodID int) (*VOD, error) {
	queryParameters := url.Values{}
	queryParameters.Add("action", vodInfoAction)
	queryParameters.Add("vod_id", strconv.Itoa(vodID))

	return decodeResponse[VOD](ctx, c, c.buildURL(apiPath, queryParameters))
}

// GetXMLTV retrieves the XMLTV EPG data.
func (c *Client) GetXMLTV(ctx context.Context) (*xmltv.EPG, error) {
	reqURL := c.buildURL(epgPath, nil)

	resp, err := c.doRequest(ctx, reqURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var epg xmltv.EPG
	if err := xml.NewDecoder(resp.Body).Decode(&epg); err != nil {
		return nil, &DecoderError{URL: redactURL(reqURL), UnderlyingDecoderError: safeCause(err)}
	}
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return nil, safeCause(err)
	}

	return &epg, nil
}

// ListLiveCategories retrieves a list of live stream categories.
func (c *Client) ListLiveCategories(ctx context.Context) ([]LiveCategory, error) {
	queryParameters := url.Values{}
	queryParameters.Add("action", liveCategoriesAction)

	reqURL := c.buildURL(apiPath, queryParameters)

	return c.listCategories(ctx, reqURL)
}

// ListLiveStreams retrieves a list of live streams.
func (c *Client) ListLiveStreams(ctx context.Context) ([]LiveStream, error) {
	queryParameters := url.Values{}
	queryParameters.Add("action", liveStreamsAction)

	reqURL := c.buildURL(apiPath, queryParameters)

	return listStreams[LiveStream](ctx, c, reqURL)
}

// ListLiveStreamsInCategory retrieves a list of live streams in a specific category.
func (c *Client) ListLiveStreamsInCategory(
	ctx context.Context,
	categoryID int,
) ([]LiveStream, error) {
	queryParameters := url.Values{}
	queryParameters.Add("action", liveStreamsAction)
	queryParameters.Add("category_id", strconv.Itoa(categoryID))

	reqURL := c.buildURL(apiPath, queryParameters)

	return listStreams[LiveStream](ctx, c, reqURL)
}

// ListSeriesCategories retrieves a list of series categories.
func (c *Client) ListSeriesCategories(ctx context.Context) ([]SeriesCategory, error) {
	queryParameters := url.Values{}
	queryParameters.Add("action", seriesCategoriesAction)

	reqURL := c.buildURL(apiPath, queryParameters)

	return c.listCategories(ctx, reqURL)
}

// ListSeriesStreams retrieves a list of series streams.
func (c *Client) ListSeriesStreams(ctx context.Context) ([]SeriesStream, error) {
	queryParameters := url.Values{}
	queryParameters.Add("action", seriesStreamsAction)

	reqURL := c.buildURL(apiPath, queryParameters)

	return listStreams[SeriesStream](ctx, c, reqURL)
}

// ListSeriesStreamsInCategory retrieves a list of series streams in a specific category.
func (c *Client) ListSeriesStreamsInCategory(
	ctx context.Context,
	categoryID int,
) ([]SeriesStream, error) {
	queryParameters := url.Values{}
	queryParameters.Add("action", seriesStreamsAction)
	queryParameters.Add("category_id", strconv.Itoa(categoryID))

	reqURL := c.buildURL(apiPath, queryParameters)

	return listStreams[SeriesStream](ctx, c, reqURL)
}

// ListVODCategories retrieves a list of VOD categories.
func (c *Client) ListVODCategories(ctx context.Context) ([]VODCategory, error) {
	queryParameters := url.Values{}
	queryParameters.Add("action", vodCategoriesAction)

	reqURL := c.buildURL(apiPath, queryParameters)

	return c.listCategories(ctx, reqURL)
}

// ListVODStreams retrieves a list of VOD streams.
func (c *Client) ListVODStreams(ctx context.Context) ([]VODStream, error) {
	queryParameters := url.Values{}
	queryParameters.Add("action", vodStreamsAction)

	reqURL := c.buildURL(apiPath, queryParameters)

	return listStreams[VODStream](ctx, c, reqURL)
}

// ListVODStreamsInCategory retrieves a list of VOD streams in a specific category.
func (c *Client) ListVODStreamsInCategory(
	ctx context.Context,
	categoryID int,
) ([]VODStream, error) {
	queryParameters := url.Values{}
	queryParameters.Add("action", vodStreamsAction)
	queryParameters.Add("category_id", strconv.Itoa(categoryID))

	reqURL := c.buildURL(apiPath, queryParameters)

	return listStreams[VODStream](ctx, c, reqURL)
}
