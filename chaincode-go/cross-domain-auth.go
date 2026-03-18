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

func (s *SmartContract) IssueTicket(ctx contractapi.TransactionContextInterface, ticketID string, targetDomain string, owner string, expiryTime int64) error {
	issuerMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get issuerMSP info: %v", err)
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

	ticketJSON, err := json.Marshal(ticket)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(ticketID, ticketJSON)
	if err != nil {
		return fmt.Errorf("failed to put state:%v", err)
	}

	err = ctx.GetStub().SetEvent("TicketIssuedEvent", ticketJSON)
	if err != nil {
		return fmt.Errorf("failed to set event:%v", err)
	}

	return nil
}

func (s *SmartContract) VerifyTicket(ctx contractapi.TransactionContextInterface, ticketID string) (bool, error) {
	ticketJSON, err := ctx.GetStub().GetState(ticketID)
	if err != nil {
		return false, fmt.Errorf("failed to read ticket:%v", err)
	}

	if ticketJSON == nil {
		return false, fmt.Errorf("ticket doesn't exist:%v", ticketID)
	}

	var ticket AuthTicket
	err = json.Unmarshal(ticketJSON, &ticket)
	if err != nil {
		return false, err
	}

	txTimeStamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return false, fmt.Errorf("failed to get time stamp:%v", err)
	}

	currentTime := txTimeStamp.Seconds
	if currentTime > ticket.ExpiryTime {
		return false, fmt.Errorf("failed to verify.Ticket has been expired at %d", ticket.ExpiryTime)
	}

	if ticket.IsRevoked {
		return false, fmt.Errorf("ticket has been revoked")
	}

	verifierMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return false, fmt.Errorf("failed to get verifier MSP ID:%v", err)
	}

	if verifierMSPID != ticket.TargetDomain {
		return false, fmt.Errorf("Permission denied")
	}
	return true, nil
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
