package homepagehandler

import (
	"log"
	"mime"
	"net/http"
	"path"
	"strings"
	"text/template"

	ssp "github.com/sqrldev/server-go-ssp"
	"github.com/sqrldev/server-go-ssp/server/homepage"
)

var sqrljs *template.Template

type TemplatedAssets struct {
	API *ssp.SqrlSspAPI
}

type jsData struct {
	RootURL string
}

// sanitizeAssetPath removes control characters from asset paths to prevent log injection
func sanitizeAssetPath(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Remove newlines, carriage returns, and other control characters
		if c == '\n' || c == '\r' || c == '\t' || c == 0x1b || c == 0x00 || c < 32 || c == 127 {
			continue
		}
		result.WriteByte(c)
	}
	return result.String()
}

// sanitizeError sanitizes error messages to prevent log injection
func sanitizeError(err error) string {
	if err == nil {
		return "(nil)"
	}
	return sanitizeAssetPath(err.Error())
}

func (ta *TemplatedAssets) Handle(w http.ResponseWriter, r *http.Request) {
	assetName := ""
	if r.URL.Path == "/" {
		assetName = "sqrl_demo.html"
	} else {
		assetName = strings.TrimLeft(r.URL.Path, "/")
	}

	if assetName == "" {
		// SECURITY: Sanitize URL path to prevent log injection
		ssp.SafeLogInfo("No asset for path: %s", sanitizeAssetPath(r.URL.Path))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	bytes, err := homepage.Asset(assetName)
	if err != nil {
		// SECURITY: Sanitize asset name to prevent log injection
		ssp.SafeLogInfo("Error getting asset %s: %s", sanitizeAssetPath(assetName), sanitizeError(err))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if assetName == "sqrlapi.js" {
		if sqrljs == nil {
			sqrljs, err = template.New("js").Parse(string(bytes))
			if err != nil {
				log.Printf("failed parsing template for sqrlapi.js: %v", err)
			}
		}
		// check again in case of error
		if sqrljs != nil {
			w.Header().Add("Content-Type", "application/javascript")
			err := sqrljs.Execute(w, jsData{ta.API.HTTPSRoot(r).String()})
			if err != nil {
				log.Printf("Failed template execute")
			}
			return
		}
	}

	ct := mime.TypeByExtension(path.Ext(assetName))
	if ct != "" {
		w.Header().Add("Content-Type", ct)
	}
	_, _ = w.Write(bytes)
}
