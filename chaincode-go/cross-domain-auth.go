package main

import (
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}
type AuthTicket struct {
	TicketID     string `json:"ticket_id"`
	IssuerDomain string `json:"issuer_domain"`
	TargetDomain string `json:"target_domain"`
	Owner        string `json:"owner"`
	ExpiryTime   int64  `json:"expiry_time"`
	IsRevoked    bool   `json:"is_revoked"`
	IssuerCert   string `json:"issuer_cert"`
}

func (s *SmartContract) getClientMSPID(ctx contractapi.TransactionContextInterface) (string, error) {
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("MSPID: %w", err)
	}
	return mspID, nil
}

func (s *SmartContract) getTicket(ctx contractapi.TransactionContextInterface, ticketID string) (*AuthTicket, error) {
	ticketJSON, err := ctx.GetStub().GetState(ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticketID: %w", err)
	}

	if ticketJSON == nil {
		return nil, fmt.Errorf("tikcet not exists: %s", ticketID)
	}

	var ticket AuthTicket
	if err := json.Unmarshal(ticketJSON, ticket); err != nil {
		return nil, fmt.Errorf("Unmarshal json: %w", err)
	}

	return &ticket, nil
}

func (s *SmartContract) putAndEvent(ctx contractapi.TransactionContextInterface, ticket *AuthTicket, eventName string) error {
	ticketJSON, err := json.Marshal(ticket)
	if err != nil {
		return fmt.Errorf("Marshal json: %w", err)
	}

	if err := ctx.GetStub().PutState(ticket.TicketID, ticketJSON); err != nil {
		return fmt.Errorf("Put state: %w", err)
	}

	if eventName != "" {
		if err := ctx.GetStub().SetEvent(eventName, ticketJSON); err != nil {
			return fmt.Errorf("Set Event: %w", err)
		}
	}
	return nil
}

func (s *SmartContract) IssueTicket(ctx contractapi.TransactionContextInterface, ticketID string, targetDomain string, owner string, expiryTime int64) error {
	issuerMSPID, err := s.getClientMSPID(ctx)
	if err != nil {
		return err
	}

	if issuerMSPID != "BusinessMSP" {
		return fmt.Errorf("Permission denied,current identity: %s", issuerMSPID)
	}

	exists, err := s.TicketExists(ctx, ticketID)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("ticket already exists:%v", ticketID)
	}

	cert, err := ctx.GetClientIdentity().GetX509Certificate()
	if err != nil {
		return fmt.Errorf("failed to get client certificate:%v", err)
	}

	certPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}))
	ticket := AuthTicket{
		TicketID:     ticketID,
		IssuerDomain: issuerMSPID,
		TargetDomain: targetDomain,
		Owner:        owner,
		ExpiryTime:   expiryTime,
		IsRevoked:    false,
		IssuerCert:   certPEM,
	}

	s.putAndEvent(ctx, &ticket, "TicketIssuedEvent")

	return nil
}

func (s *SmartContract) VerifyTicket(ctx contractapi.TransactionContextInterface, ticketID string) (bool, error) {
	ticket, err := s.getTicket(ctx, ticketID)
	if err != nil {
		return false, err
	}

	txTimeStamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return false, fmt.Errorf("failed to get time stamp:%v", err)
	}

	if txTimeStamp.Seconds > ticket.ExpiryTime {
		return false, fmt.Errorf("failed to verify.Ticket has been expired at %d", ticket.ExpiryTime)
	}

	if ticket.IsRevoked {
		return false, fmt.Errorf("ticket has been revoked")
	}

	verifierMSPID, err := s.getClientMSPID(ctx)
	if err != nil {
		return false, err
	}

	if verifierMSPID != ticket.TargetDomain {
		return false, fmt.Errorf("Permission denied")
	}
	return true, nil
}

func (s *SmartContract) RevokeTicket(ctx contractapi.TransactionContextInterface, ticketID string) error {
	callerMSPID, err := s.getClientMSPID(ctx)
	if err != nil {
		return err
	}

	if callerMSPID != "BusinessMSP" {
		return fmt.Errorf("Permission denied, current identity: %s", callerMSPID)
	}

	ticket, err := s.getTicket(ctx, ticketID)
	if err != nil {
		return err
	}

	if ticket.TargetDomain != callerMSPID {
		return fmt.Errorf("Permission denied,only issuer can revoke this ticket")
	}

	if ticket.IsRevoked == true {
		return fmt.Errorf("Ticket has been revoked")
	}

	ticket.IsRevoked = true
	return s.putAndEvent(ctx, ticket, "TicketRevokedEvent")
}

func (s *SmartContract) TicketExists(ctx contractapi.TransactionContextInterface, ticketID string) (bool, error) {
	ticketJSON, err := ctx.GetStub().GetState(ticketID)
	if err != nil {
		return false, fmt.Errorf("failed to read ticket:%v", err)
	}
	return ticketJSON != nil, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("failed to create cross domain chaincode:%v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("failed to start cross domain chaincode:%v", err)
	}
}
