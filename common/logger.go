package common

import (
	"github.com/omecodes/libome/logs"
	"net/http"
	"time"
)

type statusCatcher struct {
	status int
	w      http.ResponseWriter
}

func (catcher *statusCatcher) Header() http.Header {
	return catcher.w.Header()
}

func (catcher *statusCatcher) Write(bytes []byte) (int, error) {
	return catcher.w.Write(bytes)
}

func (catcher *statusCatcher) WriteHeader(statusCode int) {
	catcher.status = statusCode
	catcher.w.WriteHeader(statusCode)
}

func MiddlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var details []logs.NameValue
		if r.URL.RawQuery != "" {
			details = append(details, logs.Details("params", r.URL.RawQuery))
		}
		details = append(details, logs.Details("time", start))

		logs.Info(r.Method+" "+r.RequestURI, details...)

		c := &statusCatcher{
			status: 0,
			w:      w,
		}

		next.ServeHTTP(c, r)
		duration := time.Since(start)

		if c.status == http.StatusOK || c.status == 0 {
			logs.Info(r.Method+" "+r.RequestURI,
				logs.Details("duration", duration.String()),
			)
		} else {
			logs.Error(r.Method+" "+r.RequestURI+" "+http.StatusText(c.status),
				logs.Details("duration", duration.String()),
			)
		}
	})
}
