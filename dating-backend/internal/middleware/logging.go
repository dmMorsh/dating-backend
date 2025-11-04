package middleware

import (
	"log"
	"net/http"
)

// LoggingMiddleware logs each incoming HTTP request and then calls the next
// handler. This is a minimal middleware suitable for development and
// lightweight production logging.
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Printf("[%s] %s %s", r.RemoteAddr, r.Method, r.URL.Path)
        next(w, r)
    }
}

// package middleware

// import (
// 	"bytes"
// 	"io"
// 	"log"
// 	"net/http"
// )

// // обёртка вокруг ResponseWriter, чтобы перехватывать ответ
// type responseRecorder struct {
// 	http.ResponseWriter
// 	statusCode int
// 	body       bytes.Buffer
// }

// func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
// 	return &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
// }

// // перехватываем запись в тело
// func (rr *responseRecorder) Write(b []byte) (int, error) {
// 	rr.body.Write(b) // сохраняем копию
// 	return rr.ResponseWriter.Write(b)
// }

// // перехватываем установку статуса
// func (rr *responseRecorder) WriteHeader(statusCode int) {
// 	rr.statusCode = statusCode
// 	rr.ResponseWriter.WriteHeader(statusCode)
// }

// // основное middleware
// func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		rec := newResponseRecorder(w)

// 		// читаем тело запроса (если нужно)
// 		var requestBody []byte
// 		if r.Body != nil {
// 			requestBody, _ = io.ReadAll(r.Body)
// 			r.Body = io.NopCloser(bytes.NewBuffer(requestBody)) // возвращаем тело обратно, чтобы хэндлер мог его прочитать
// 		}

// 		// выполняем обработчик
// 		next(rec, r)

// 		// логируем
// 		log.Printf("[%s] %s %s (status=%d)", r.RemoteAddr, r.Method, r.URL.Path, rec.statusCode)
// 		if len(requestBody) > 0 {
// 			log.Printf("→ Request body: %s", string(requestBody))
// 		}
// 		if rec.body.Len() > 0 {
// 			log.Printf("← Response: %s", rec.body.String())
// 		}
// 	}
// }
