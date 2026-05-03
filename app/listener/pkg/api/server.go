package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	collection *mongo.Collection
}

func NewServer(col *mongo.Collection) *Server {
	return &Server{collection: col}
}

func (s *Server) Start(port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/verify", s.handleVerify)

	log.Printf("Starting API server on port %s", port)
	return http.ListenAndServe(":"+port, mux)
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ticketID := r.URL.Query().Get("ticketId")
	if ticketID == "" {
		http.Error(w, "Missing ticketId", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var ticket struct {
		TicketID     string `bson:"ticket_id"`
		TargetDomain string `bson:"target_domain"`
		IsRevoked    bool   `bson:"is_revoked"`
		ExpiryTime   int64  `bson:"expiry_time"`
	}

	err := s.collection.FindOne(ctx, bson.M{"_id": ticketID}).Decode(&ticket)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "凭证不存在",
			})
			return
		}
		log.Printf("Database error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	currentTime := time.Now().Unix()

	if ticket.IsRevoked {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "凭证已被撤销",
		})
		return
	}

	if ticket.ExpiryTime < currentTime {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "凭证已过期",
		})
		return
	}

	if ticket.TargetDomain != "PrivacyMSP" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "凭证目标域不匹配",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "验证成功，允许登录",
	})
}
