package xtreamcodes

import "github.com/sherif-fanous/xtreamcodes/types"

// The canonical model definitions live in the types package so that consumers
// who only need serialization can import them without pulling in the HTTP
// client and its dependencies. These aliases keep the models available on the
// root package for client users.
type (
	AuthInfo   = types.AuthInfo
	UserInfo   = types.UserInfo
	ServerInfo = types.ServerInfo

	Category       = types.Category
	LiveCategory   = types.LiveCategory
	SeriesCategory = types.SeriesCategory
	VODCategory    = types.VODCategory

	EPG        = types.EPG
	EPGListing = types.EPGListing

	LiveStream = types.LiveStream

	Series       = types.Series
	SeriesStream = types.SeriesStream
	SeriesInfo   = types.SeriesInfo
	Episode      = types.Episode
	Season       = types.Season

	VOD       = types.VOD
	VODInfo   = types.VODInfo
	VODStream = types.VODStream
	MovieData = types.MovieData
)
