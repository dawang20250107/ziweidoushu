package httpapi

import (
	"encoding/json"
	"net/http"
)

// envelope 统一响应包裹:{"ok":true,"data":...} / {"ok":false,"error":{...}}。
type envelope struct {
	OK    bool      `json:"ok"`
	Data  any       `json:"data,omitempty"`
	Error *apiError `json:"error,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{OK: status < 400, Data: data})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{OK: false, Error: &apiError{Code: code, Message: message}})
}

// maxBodyBytes 请求体上限:所有 JSON 端点的正常载荷远小于 1MB(最大是
// 问星对话的多轮历史),超限直接 4xx,防止超大请求体拖垮解码与内存。
const maxBodyBytes = 1 << 20

// decodeJSON 解析请求体,出错时直接写响应并返回 false。
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "请求体解析失败: "+err.Error())
		return false
	}
	return true
}
