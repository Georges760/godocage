package docage

import (
	"bytes"
	"github.com/jung-kurt/gofpdf"
	"log"
	"os"
	"testing"
)

var api API

func TestMain(m *testing.M) {
	// api.Email = "YourAccountEmail"
	// api.Key = "YourAccountAPIKey"
	var err error
	var contactid string
	var contact Contact
	contact.Email = "moi@free.fr"
	contact.FirstName = "John"
	contact.LastName = "Doe"
	contact.Country = "France"
	contactid, err = api.CreateContact(contact)
	if err != nil {
		log.Fatal("Error creating contact: ", err)
	}
	log.Println("contactid:", contactid)

	var txid string
	var tx Transaction
	tx.Name = "Test Tx"
	tx.IsTest = true
	txid, err = api.CreateTransaction(tx)
	if err != nil {
		log.Fatal("Error creating transaction: ", err)
	}
	log.Println("txid:", txid)

	var status string
	status, err = api.GetTransactionStatus(txid)
	if err != nil {
		log.Fatal("Error getting transaction status: ", err)
	}
	log.Println("status:", status)

	var memberid string
	var member TransactionMember
	member.TransactionId = txid
	member.ContactId = contactid
	member.SignMode = 1 // Email
	memberid, err = api.AddTransactionMember(member)
	if err != nil {
		log.Fatal("Error adding member: ", err)
	}
	log.Println("memberid:", memberid)

	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Hello, world")
	var file bytes.Buffer
	err = pdf.Output(&file)
	if err != nil {
		log.Fatal("Error creating test PDF: ", err)
	}
	err = api.AddTransactionDocument(txid, "document.pdf", &file)
	if err != nil {
		log.Fatal("Error adding document: ", err)
	}

	os.Exit(m.Run())
}
