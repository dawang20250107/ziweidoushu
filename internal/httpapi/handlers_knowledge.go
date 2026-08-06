package httpapi

import (
	"encoding/json"
	"net/http"
)

// handleNihai 三纪知识库透传(tianji | renji | diji | bio)。
func (s *Server) handleNihai(w http.ResponseWriter, r *http.Request) {
	var raw json.RawMessage
	switch r.PathValue("section") {
	case "tianji":
		raw = s.kb.Tianji
	case "renji":
		raw = s.kb.Renji
	case "diji":
		raw = s.kb.Diji
	case "bio":
		raw = s.kb.Bio
	default:
		writeError(w, http.StatusNotFound, "not_found", "支持的分区:tianji / renji / diji / bio")
		return
	}
	writeJSON(w, http.StatusOK, raw)
}

func (s *Server) handleKnowledgeStars(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"descriptions": s.kb.StarDesc,
		"slugs":        s.kb.StarSlugs,
		"order":        s.kb.StarOrder,
		// 全量星曜档案层:lore 逐星义理、cycles 四大十二神逐名义、flow 流曜十义
		"lore":   s.kb.StarLore,
		"cycles": s.kb.StarCycles,
		"flow":   s.kb.StarFlow,
	})
}

func (s *Server) handleKnowledgeTopics(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.kb.Topics)
}

func (s *Server) handleKnowledgeHeming(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.kb.Heming)
}

func (s *Server) handleCities(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"provinces": s.kb.Provinces})
}
