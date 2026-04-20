package httph

import "net/http"

func SendRaw(w http.ResponseWriter, statusCode int, mimeType string, data []byte) {
	if mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	}

	w.WriteHeader(statusCode)

	if len(data) > 0 {
		_, _ = w.Write(data)
	}
}

func SendEmpty(w http.ResponseWriter, statusCode int) {
	SendRaw(w, statusCode, "", nil)
}

func SendEncodedWithMIME(w http.ResponseWriter, r *http.Request, statusCode int, mimeType string, obj any) {
	if mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	}

	if err := EncodeJSON(w, obj); err != nil {
		ErrorApply(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(statusCode)
}

func SendEncoded(w http.ResponseWriter, r *http.Request, statusCode int, obj any) {
	SendEncodedWithMIME(w, r, statusCode, MIMEApplicationJSONCharsetUTF8, obj)
}
