package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrTikcetNotFound = errors.New("ticket not found")
	ErrTikcetRevoked  = errors.New("ticket revoked")
	ErrTikcetExpired  = errors.New("ticket expired")
	ErrDomainMismatch = errors.New("domain mismatch")
)

type (
	Server struct {
		collection *mongo.Collection
		mspID      string
		httpSever  *http.Server
	}
	Ticket struct {
		TicketID     string `bson:"ticket_id"`
		TargetDomain string `bson:"target_domain"`
		IsRevoked    bool   `bson:"is_revoked"`
		ExpiryTime   int64  `bson:"expiry_time"`
	}
)

func NewServer(col *mongo.Collection, mspID string) *Server {
	return &Server{
		collection: col,
		mspID:      mspID,
	}
}

func (s *Server) Start(port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/verify", s.handleVerify)
	s.httpSever = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Starting API server on port %s", port)
	return s.httpSever.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpSever != nil {
		log.Printf("Shutting down http server")
		return s.httpSever.Shutdown(ctx)
	}
	return nil
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, false, "method not allowed")
		return
	}

	ticketID := r.URL.Query().Get("ticketId")
	if ticketID == "" {
		writeJSON(w, http.StatusBadRequest, false, "missing parameter")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := s.validateTicket(ctx, ticketID)
	if err != nil {
		handleErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, true, "Verified")
}

func writeJSON(w http.ResponseWriter, statusCode int, success bool, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]any{
		"success": success,
		"message": message,
	})
}

func handleErr(w http.ResponseWriter, err error) {
	switch err {
	case ErrTikcetNotFound, ErrTikcetRevoked, ErrTikcetExpired, ErrDomainMismatch:
		writeJSON(w, http.StatusUnauthorized, false, err.Error())
	default:
		log.Printf("Internal Sever error: %vr", err)
		writeJSON(w, http.StatusInternalServerError, false, "internal server error")
	}
}

func (s *Server) validateTicket(ctx context.Context, ticketID string) error {
	ticket, err := s.getTicket(ctx, ticketID)
	now := time.Now().Unix()
	if err != nil {
		return err
	}

	if ticket.IsRevoked {
		return ErrTikcetRevoked
	}

	if ticket.TargetDomain != s.mspID {
		return ErrDomainMismatch
	}

	if ticket.ExpiryTime < now {
		return ErrTikcetExpired
	}

	return nil
}

func (s *Server) getTicket(ctx context.Context, ticketID string) (*Ticket, error) {
	var ticket Ticket
	err := s.collection.FindOne(ctx, bson.M{"_id": ticketID}).Decode(&ticket)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrTikcetNotFound
		}
		return nil, err
	}

	return &ticket, err
}
