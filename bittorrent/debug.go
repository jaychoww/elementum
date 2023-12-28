package bittorrent

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/anacrolix/missinggo/perf"
	"github.com/gin-gonic/gin"

	"github.com/elgatito/elementum/config"
	"github.com/elgatito/elementum/util/ip"
	"github.com/elgatito/elementum/xbmc"
)

func DebugAll(s *Service) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer perf.ScopeTimer()()

		ctx.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

		writeHeader(ctx.Writer, "Torrent Client")
		writeResponse(ctx.Writer, "/info")

		writeHeader(ctx.Writer, "Debug Perf")
		writeResponse(ctx.Writer, "/debug/perf")

		writeHeader(ctx.Writer, "Debug LockTimes")
		writeResponse(ctx.Writer, "/debug/lockTimes")

		writeHeader(ctx.Writer, "Debug GoRoutines")
		writeResponse(ctx.Writer, "/debug/pprof/goroutine?debug=1")

		writeHeader(ctx.Writer, "Debug Vars")
		writeResponse(ctx.Writer, "/debug/vars")
	}
}

// DebugBundle ...
func DebugBundle(s *Service) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer perf.ScopeTimer()()

		xbmcHost, err := xbmc.GetXBMCHostWithContext(ctx)
		if err != nil {
			log.Infof("Could not find attached Kodi: %s", err)
			return
		}

		var logFile io.Reader

		logPath := xbmcHost.TranslatePath("special://logpath/kodi.log")
		logFile, err = os.Open(logPath)
		if err != nil {
			log.Debug("Could not open kodi.log locally, trying to get it from Kodi")
			logFile = bytes.NewReader(xbmcHost.GetKodiLog())
		} else {
			defer logFile.(*os.File).Close()
		}

		now := time.Now()
		fileName := fmt.Sprintf("bundle_%d_%d_%d_%d_%d.log", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())
		ctx.Writer.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		ctx.Writer.Header().Set("Content-Type", "text/plain")

		writeHeader(ctx.Writer, "Torrent Client")
		writeResponse(ctx.Writer, "/info")

		writeHeader(ctx.Writer, "Debug Perf")
		writeResponse(ctx.Writer, "/debug/perf")

		writeHeader(ctx.Writer, "Debug LockTimes")
		writeResponse(ctx.Writer, "/debug/lockTimes")

		writeHeader(ctx.Writer, "Debug GoRoutines")
		writeResponse(ctx.Writer, "/debug/pprof/goroutine?debug=1")

		writeHeader(ctx.Writer, "Debug Vars")
		writeResponse(ctx.Writer, "/debug/vars")

		writeHeader(ctx.Writer, "kodi.log")
		io.Copy(ctx.Writer, logFile)
	}
}

func writeHeader(w http.ResponseWriter, title string) {
	w.Write([]byte("\n\n" + strings.Repeat("-", 70) + "\n"))
	w.Write([]byte(title))
	w.Write([]byte("\n" + strings.Repeat("-", 70) + "\n\n"))
}

func writeResponse(w http.ResponseWriter, url string) {
	w.Write([]byte("Response for url: " + url + "\n\n"))

	resp, err := http.Get(fmt.Sprintf("http://%s:%d%s", ip.GetLocalHost(), config.Args.LocalPort, url))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	io.Copy(w, resp.Body)
}
