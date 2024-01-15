package trakt

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/elgatito/elementum/cache"
	"github.com/elgatito/elementum/config"
	"github.com/elgatito/elementum/library/playcount"
	"github.com/elgatito/elementum/library/uid"
	"github.com/elgatito/elementum/tmdb"
	"github.com/elgatito/elementum/util"
	"github.com/elgatito/elementum/util/reqapi"
	"github.com/elgatito/elementum/xbmc"

	"github.com/anacrolix/missinggo/perf"
	"github.com/anacrolix/sync"
	"github.com/jmcvetta/napping"
)

// Fill fanart from TMDB
func setFanart(movie *Movie) *Movie {
	if movie.Images == nil {
		movie.Images = &Images{}
	}
	if movie.Images.Poster == nil {
		movie.Images.Poster = &Sizes{}
	}
	if movie.Images.Thumbnail == nil {
		movie.Images.Thumbnail = &Sizes{}
	}
	if movie.Images.FanArt == nil {
		movie.Images.FanArt = &Sizes{}
	}
	if movie.Images.Banner == nil {
		movie.Images.Banner = &Sizes{}
	}
	if movie.Images.ClearArt == nil {
		movie.Images.ClearArt = &Sizes{}
	}

	if movie.IDs == nil || movie.IDs.TMDB == 0 {
		return movie
	}

	tmdbImages := tmdb.GetImages(movie.IDs.TMDB)
	if tmdbImages == nil {
		return movie
	}

	if len(tmdbImages.Posters) > 0 {
		posterImage := tmdb.ImageURL(tmdbImages.Posters[0].FilePath, "w1280")
		for _, image := range tmdbImages.Posters {
			if image.Iso639_1 == config.Get().Language {
				posterImage = tmdb.ImageURL(image.FilePath, "w1280")
				break
			}
		}
		movie.Images.Poster.Full = posterImage
		movie.Images.Thumbnail.Full = posterImage
	}
	if len(tmdbImages.Backdrops) > 0 {
		backdropImage := tmdb.ImageURL(tmdbImages.Backdrops[0].FilePath, "w1280")
		for _, image := range tmdbImages.Backdrops {
			if image.Iso639_1 == config.Get().Language {
				backdropImage = tmdb.ImageURL(image.FilePath, "w1280")
				break
			}
		}
		movie.Images.FanArt.Full = backdropImage
		movie.Images.Banner.Full = backdropImage
	}
	return movie
}

func setFanarts(movies []*Movies) []*Movies {
	wg := sync.WaitGroup{}
	for i, movie := range movies {
		wg.Add(1)
		go func(idx int, m *Movies) {
			defer wg.Done()
			movies[idx].Movie = setFanart(m.Movie)
		}(i, movie)
	}
	wg.Wait()

	return movies
}

func setCalendarFanarts(movies []*CalendarMovie) []*CalendarMovie {
	wg := sync.WaitGroup{}
	for i, movie := range movies {
		wg.Add(1)
		go func(idx int, m *CalendarMovie) {
			defer wg.Done()
			movies[idx].Movie = setFanart(m.Movie)
		}(i, movie)
	}
	wg.Wait()

	return movies
}

// GetMovie ...
func GetMovie(ID string) (movie *Movie) {
	defer perf.ScopeTimer()()

	cacheStore := cache.NewDBStore()
	key := fmt.Sprintf(cache.TraktMovieKey, ID)
	if err := cacheStore.Get(key, &movie); err != nil {
		req := reqapi.Request{
			API:    reqapi.TraktAPI,
			URL:    fmt.Sprintf("movies/%s", ID),
			Header: GetAvailableHeader(),
			Params: napping.Params{
				"extended": "full,images",
			}.AsUrlValues(),
			Result: &movie,
		}

		if err = req.Do(); err != nil {
			log.Error(err)
			if xbmcHost, _ := xbmc.GetLocalXBMCHost(); xbmcHost != nil {
				xbmcHost.Notify("Elementum", fmt.Sprintf("Failed getting Trakt movie (%s), check your logs.", ID), config.AddonIcon())
			}
			return
		}

		cacheStore.Set(key, movie, cache.TraktMovieExpire)
	}

	return
}

// GetMovieByTMDB ...
func GetMovieByTMDB(tmdbID string) (movie *Movie) {
	defer perf.ScopeTimer()()

	cacheStore := cache.NewDBStore()
	key := fmt.Sprintf(cache.TraktMovieByTMDBKey, tmdbID)
	if err := cacheStore.Get(key, &movie); err != nil {
		var results MovieSearchResults
		req := reqapi.Request{
			API:    reqapi.TraktAPI,
			URL:    fmt.Sprintf("search/tmdb/%s?type=movie", tmdbID),
			Header: GetAvailableHeader(),
			Params: napping.Params{}.AsUrlValues(),
			Result: &results,
		}

		if err = req.Do(); err != nil {
			log.Error(err)
			if xbmcHost, _ := xbmc.GetLocalXBMCHost(); xbmcHost != nil {
				xbmcHost.Notify("Elementum", "Failed getting Trakt movie using TMDB ID, check your logs.", config.AddonIcon())
			}
			return
		}

		if len(results) > 0 && results[0].Movie != nil {
			movie = results[0].Movie
		}
		cacheStore.Set(key, movie, cache.TraktMovieByTMDBExpire)
	}
	return
}

// SearchMovies ...
func SearchMovies(query string, page string) (movies []*Movies, err error) {
	defer perf.ScopeTimer()()

	req := &reqapi.Request{
		API:    reqapi.TraktAPI,
		URL:    "search",
		Header: GetAvailableHeader(),
		Params: napping.Params{
			"page":     page,
			"limit":    strconv.Itoa(config.Get().ResultsPerPage),
			"query":    query,
			"extended": "full,images",
		}.AsUrlValues(),
		Result: &movies,
	}

	if err = req.Do(); err != nil {
		return
	}

	// TODO use response headers for pagination limits:
	// X-Pagination-Page-Count:10
	// X-Pagination-Item-Count:100

	return
}

// TopMovies ...
func TopMovies(topCategory string, page string) (movies []*Movies, total int, err error) {
	defer perf.ScopeTimer()()

	endPoint := "movies/" + topCategory
	if topCategory == "recommendations" {
		endPoint = topCategory + "/movies"
	}

	resultsPerPage := config.Get().ResultsPerPage
	limit := resultsPerPage
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return
	}
	if pageInt < -1 {
		resultsPerPage = pageInt * -1
		limit = pageInt * -1
		page = "1"
		pageInt = 1
	}
	page = strconv.Itoa((pageInt-1)*resultsPerPage/limit + 1)

	cacheStore := cache.NewDBStore()
	key := fmt.Sprintf(cache.TraktMoviesByCategoryKey, topCategory, page, limit)
	totalKey := fmt.Sprintf(cache.TraktMoviesByCategoryTotalKey, topCategory)
	if err := cacheStore.Get(key, &movies); err != nil || len(movies) == 0 {
		var movieList []*Movie
		req := &reqapi.Request{
			API:    reqapi.TraktAPI,
			URL:    endPoint,
			Header: GetAvailableHeader(),
			Params: napping.Params{
				"page":     page,
				"limit":    strconv.Itoa(limit),
				"extended": "full,images",
			}.AsUrlValues(),
			Result: &movies,
		}

		if topCategory == "popular" || topCategory == "recommendations" {
			req.Result = &movieList
		}

		if err = req.Do(); err != nil {
			return movies, 0, err
		}

		if topCategory == "popular" || topCategory == "recommendations" {
			movieListing := make([]*Movies, 0)
			for _, movie := range movieList {
				movieItem := Movies{
					Movie: movie,
				}
				movieListing = append(movieListing, &movieItem)
			}
			movies = movieListing
		}

		pagination := getPagination(req.ResponseHeader)
		total = pagination.ItemCount
		if err != nil {
			log.Warning(err)
		} else {
			cacheStore.Set(totalKey, total, cache.TraktMoviesByCategoryTotalExpire)
		}

		cacheStore.Set(key, movies, cache.TraktMoviesByCategoryExpire)
	} else {
		if err := cacheStore.Get(totalKey, &total); err != nil {
			total = -1
		}
	}

	return
}

// PreviousWatchlistMovies ...
func PreviousWatchlistMovies() (movies []*Movies, err error) {
	err = cache.
		NewDBStore().
		Get(cache.TraktMoviesWatchlistKey, &movies)

	return movies, err
}

// WatchlistMovies ...
func WatchlistMovies(isUpdateNeeded bool) (movies []*Movies, err error) {
	if err := Authorized(); err != nil {
		return movies, err
	}

	defer perf.ScopeTimer()()

	cacheStore := cache.NewDBStore()

	if !isUpdateNeeded {
		if err := cacheStore.Get(cache.TraktMoviesWatchlistKey, &movies); err == nil {
			return movies, nil
		}
	}

	var watchlist []*WatchlistMovie
	req := &reqapi.Request{
		API:    reqapi.TraktAPI,
		URL:    "sync/watchlist/movies",
		Header: GetAvailableHeader(),
		Params: napping.Params{
			"extended": "full,images",
		}.AsUrlValues(),
		Result: &watchlist,
	}

	if err = req.Do(); err != nil {
		return movies, err
	}

	movieListing := make([]*Movies, 0)
	for _, movie := range watchlist {
		movieItem := Movies{
			Movie: movie.Movie,
		}
		movieListing = append(movieListing, &movieItem)
	}
	movies = movieListing

	cacheStore.Set(cache.TraktMoviesWatchlistKey, &movies, cache.TraktMoviesWatchlistExpire)
	return
}

// PreviousCollectionMovies ...
func PreviousCollectionMovies() (movies []*Movies, err error) {
	err = cache.
		NewDBStore().
		Get(cache.TraktMoviesCollectionKey, &movies)

	return movies, err
}

// CollectionMovies ...
func CollectionMovies(isUpdateNeeded bool) (movies []*Movies, err error) {
	if errAuth := Authorized(); errAuth != nil {
		return movies, errAuth
	}

	defer perf.ScopeTimer()()

	cacheStore := cache.NewDBStore()

	if !isUpdateNeeded {
		if err := cacheStore.Get(cache.TraktMoviesCollectionKey, &movies); err == nil {
			return movies, nil
		}
	}

	var collection []*CollectionMovie
	req := &reqapi.Request{
		API:    reqapi.TraktAPI,
		URL:    "sync/collection/movies",
		Header: GetAvailableHeader(),
		Params: napping.Params{
			"extended": "full,images",
		}.AsUrlValues(),
		Result: &collection,
	}

	if err = req.Do(); err != nil {
		return movies, err
	}

	movieListing := make([]*Movies, 0)
	for _, movie := range collection {
		movieItem := Movies{
			Movie: movie.Movie,
		}
		movieListing = append(movieListing, &movieItem)
	}
	movies = movieListing

	cacheStore.Set(cache.TraktMoviesCollectionKey, &movies, cache.TraktMoviesCollectionExpire)
	return movies, err
}

// Userlists ...
func Userlists() (lists []*List) {
	defer perf.ScopeTimer()()

	traktUsername := config.Get().TraktUsername
	if traktUsername == "" || config.Get().TraktToken == "" || !config.Get().TraktAuthorized {
		if xbmcHost, _ := xbmc.GetLocalXBMCHost(); xbmcHost != nil {
			xbmcHost.Notify("Elementum", "LOCALIZE[30149]", config.AddonIcon())
		}
		return lists
	}
	endPoint := fmt.Sprintf("users/%s/lists", traktUsername)

	req := &reqapi.Request{
		API:    reqapi.TraktAPI,
		URL:    endPoint,
		Header: GetAvailableHeader(),
		Params: napping.Params{}.AsUrlValues(),
		Result: &lists,
	}

	if err := req.Do(); err != nil {
		if xbmcHost, _ := xbmc.GetLocalXBMCHost(); xbmcHost != nil {
			xbmcHost.Notify("Elementum", err.Error(), config.AddonIcon())
		}
		log.Error(err)
		return lists
	}

	sort.Slice(lists, func(i int, j int) bool {
		return lists[i].Name < lists[j].Name
	})

	return lists
}

// Likedlists ...
func Likedlists() (lists []*List) {
	defer perf.ScopeTimer()()

	traktUsername := config.Get().TraktUsername
	if traktUsername == "" || config.Get().TraktToken == "" {
		if xbmcHost, _ := xbmc.GetLocalXBMCHost(); xbmcHost != nil {
			xbmcHost.Notify("Elementum", "LOCALIZE[30149]", config.AddonIcon())
		}
		return lists
	}

	inputLists := []*ListContainer{}
	req := &reqapi.Request{
		API:    reqapi.TraktAPI,
		URL:    "users/likes/lists",
		Header: GetAvailableHeader(),
		Params: napping.Params{}.AsUrlValues(),
		Result: &inputLists,
	}

	if err := req.Do(); err != nil {
		if xbmcHost, _ := xbmc.GetLocalXBMCHost(); xbmcHost != nil {
			xbmcHost.Notify("Elementum", err.Error(), config.AddonIcon())
		}
		log.Error(err)
		return lists
	}

	for _, l := range inputLists {
		lists = append(lists, l.List)
	}

	sort.Slice(lists, func(i int, j int) bool {
		return lists[i].Name < lists[j].Name
	})

	return lists
}

// TopLists ...
func TopLists(page string) (lists []*ListContainer, hasNext bool) {
	defer perf.ScopeTimer()()

	pageInt, _ := strconv.Atoi(page)

	req := &reqapi.Request{
		API:    reqapi.TraktAPI,
		URL:    "lists/popular",
		Header: GetAvailableHeader(),
		Params: napping.Params{
			"page":  page,
			"limit": strconv.Itoa(ListsPerPage),
		}.AsUrlValues(),
		Result: &lists,
	}

	if err := req.Do(); err != nil {
		if xbmcHost, _ := xbmc.GetLocalXBMCHost(); xbmcHost != nil {
			xbmcHost.Notify("Elementum", err.Error(), config.AddonIcon())
		}
		log.Error(err)
		return lists, hasNext
	}

	p := getPagination(req.ResponseHeader)
	hasNext = p.PageCount > pageInt

	return lists, hasNext
}

// PreviousListItemsMovies ...
func PreviousListItemsMovies(listID string) (movies []*Movies, err error) {
	cacheStore := cache.NewDBStore()
	key := fmt.Sprintf(cache.TraktMoviesListKey, listID)
	err = cacheStore.Get(key, &movies)

	return
}

// ListItemsMovies ...
func ListItemsMovies(user string, listID string, isUpdateNeeded bool) (movies []*Movies, err error) {
	defer perf.ScopeTimer()()

	if user == "" || user == "id" {
		user = config.Get().TraktUsername
	}

	cacheStore := cache.NewDBStore()
	key := fmt.Sprintf(cache.TraktMoviesListKey, listID)

	if !isUpdateNeeded {
		if err := cacheStore.Get(key, &movies); err == nil {
			return movies, nil
		}
	}

	var list []*ListItem
	req := &reqapi.Request{
		API:    reqapi.TraktAPI,
		URL:    fmt.Sprintf("users/%s/lists/%s/items/movies", user, listID),
		Header: GetAvailableHeader(),
		Params: napping.Params{}.AsUrlValues(),
		Result: &list,
	}

	if err = req.Do(); err != nil {
		return movies, err
	}

	movieListing := make([]*Movies, 0)
	for _, movie := range list {
		if movie.Movie == nil {
			continue
		}
		movieItem := Movies{
			Movie: movie.Movie,
		}
		movieListing = append(movieListing, &movieItem)
	}
	movies = movieListing

	cacheStore.Set(key, &movies, 1*time.Minute)
	return movies, err
}

// CalendarMovies ...
func CalendarMovies(endPoint string, page string) (movies []*CalendarMovie, total int, err error) {
	defer perf.ScopeTimer()()

	resultsPerPage := config.Get().ResultsPerPage
	limit := resultsPerPage
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return
	}
	page = strconv.Itoa((pageInt-1)*resultsPerPage/limit + 1)

	cacheStore := cache.NewDBStore()
	endPointKey := strings.Replace(endPoint, "/", ".", -1)
	key := fmt.Sprintf(cache.TraktMoviesCalendarKey, endPointKey, page, limit)
	totalKey := fmt.Sprintf(cache.TraktMoviesCalendarTotalKey, endPointKey)
	if err := cacheStore.Get(key, &movies); err != nil {
		req := &reqapi.Request{
			API:    reqapi.TraktAPI,
			URL:    "calendars/" + endPoint,
			Header: GetAuthenticatedHeader(),
			Params: napping.Params{
				"page":     page,
				"limit":    strconv.Itoa(limit),
				"extended": "full,images",
			}.AsUrlValues(),
			Result: &movies,
		}

		if err = req.Do(); err != nil {
			log.Error(err)
			return movies, 0, err
		}

		pagination := getPagination(req.ResponseHeader)
		total = pagination.ItemCount
		if err != nil {
			total = -1
		} else {
			cacheStore.Set(totalKey, total, cache.TraktMoviesCalendarTotalExpire)
		}

		cacheStore.Set(key, &movies, cache.TraktMoviesCalendarExpire)
	} else {
		if err := cacheStore.Get(totalKey, &total); err != nil {
			total = -1
		}
	}

	return
}

// WatchedMovies ...
func WatchedMovies(isUpdateNeeded bool) ([]*WatchedMovie, error) {
	defer perf.ScopeTimer()()

	var movies []*WatchedMovie
	err := Request(
		"sync/watched/movies",
		napping.Params{},
		true,
		isUpdateNeeded,
		cache.TraktMoviesWatchedKey,
		cache.TraktMoviesWatchedExpire,
		&movies,
	)

	sort.Slice(movies, func(i int, j int) bool {
		return movies[i].LastWatchedAt.Unix() > movies[j].LastWatchedAt.Unix()
	})

	if len(movies) != 0 {
		cache.
			NewDBStore().
			Set(cache.TraktMoviesWatchedKey, &movies, cache.TraktMoviesWatchedExpire)
	}

	return movies, err
}

// PreviousWatchedMovies ...
func PreviousWatchedMovies() (movies []*WatchedMovie, err error) {
	err = cache.
		NewDBStore().
		Get(cache.TraktMoviesWatchedKey, &movies)

	return
}

// PausedMovies ...
func PausedMovies(isUpdateNeeded bool) ([]*PausedMovie, error) {
	defer perf.ScopeTimer()()

	var movies []*PausedMovie
	err := Request(
		"sync/playback/movies",
		napping.Params{
			"extended": "full",
		},
		true,
		isUpdateNeeded,
		cache.TraktMoviesPausedKey,
		cache.TraktMoviesPausedExpire,
		&movies,
	)

	return movies, err
}

// ToListItem ...
func (movie *Movie) ToListItem() (item *xbmc.ListItem) {
	defer perf.ScopeTimer()()

	if !config.Get().ForceUseTrakt && movie.IDs.TMDB != 0 {
		tmdbID := strconv.Itoa(movie.IDs.TMDB)
		if tmdbMovie := tmdb.GetMovieByID(tmdbID, config.Get().Language); tmdbMovie != nil {
			item = tmdbMovie.ToListItem()
		}
	}
	if item == nil {
		movie = setFanart(movie)
		item = &xbmc.ListItem{
			Label: movie.Title,
			Info: &xbmc.ListItemInfo{
				Count:         rand.Int(),
				Title:         movie.Title,
				OriginalTitle: movie.Title,
				Year:          movie.Year,
				Genre:         movie.Genres,
				Plot:          movie.Overview,
				PlotOutline:   movie.Overview,
				TagLine:       movie.TagLine,
				Rating:        movie.Rating,
				Votes:         strconv.Itoa(movie.Votes),
				Duration:      movie.Runtime * 60,
				MPAA:          movie.Certification,
				Code:          movie.IDs.IMDB,
				IMDBNumber:    movie.IDs.IMDB,
				Trailer:       util.TrailerURL(movie.Trailer),
				PlayCount:     playcount.GetWatchedMovieByTMDB(movie.IDs.TMDB).Int(),
				DBTYPE:        "movie",
				Mediatype:     "movie",
			},
			Art: &xbmc.ListItemArt{
				Poster:    movie.Images.Poster.Full,
				FanArt:    movie.Images.FanArt.Full,
				Banner:    movie.Images.Banner.Full,
				Thumbnail: movie.Images.Thumbnail.Full,
				ClearArt:  movie.Images.ClearArt.Full,
			},
			Thumbnail: movie.Images.Poster.Full,
			UniqueIDs: &xbmc.UniqueIDs{
				TMDB: strconv.Itoa(movie.IDs.TMDB),
			},
		}
	}

	if lm, err := uid.GetMovieByTMDB(movie.IDs.TMDB); lm != nil && err == nil {
		item.Info.DBID = lm.UIDs.Kodi
	}

	if len(item.Info.Trailer) == 0 {
		item.Info.Trailer = util.TrailerURL(movie.Trailer)
	}

	return
}
