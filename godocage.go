package docage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

const baseURL = "https://api.docage.com/"

type API struct {
	Email string
	Key   string
}

type Contact struct {
	Email                string `json:"Email,omitempty"`
	FirstName            string `json:"FirstName,omitempty"`
	LastName             string `json:"LastName,omitempty"`
	Address1             string `json:"Address1,omitempty"`
	Address2             string `json:"Address2,omitempty"`
	City                 string `json:"City,omitempty"`
	State                string `json:"State,omitempty"`
	ZipCode              string `json:"ZipCode,omitempty"`
	Country              string `json:"Country,omitempty"`
	Notes                string `json:"Notes,omitempty"`
	Phone                string `json:"Phone,omitempty"`
	Mobile               string `json:"Mobile,omitempty"`
	Company              string `json:"Company,omitempty"`
	Gender               string `json:"Gender,omitempty"`
	Civility             string `json:"Civility,omitempty"`
	ProfilePictureSmall  string `json:"ProfilePictureSmall,omitempty"`
	ProfilePictureMedium string `json:"ProfilePictureMedium,omitempty"`
	ProfilePictureLarge  string `json:"ProfilePictureLarge,omitempty"`
}

type TransactionMember struct {
	TransactionId    string `json:"TransactionId,omitempty"`
	ContactId        string `json:"ContactId,omitempty"`
	NotifyInvitation bool   `json:"NotifyInvitation,omitempty"` // default: true
	NotifySignature  bool   `json:"NotifySignature,omitempty"`  // default: true
	NotifyRefusal    bool   `json:"NotifyRefusal,omitempty"`    // default: true
	NotifyCompletion bool   `json:"NotifyCompletion,omitempty"` // default: true
	MemberRole       int    `json:"MemberRole,omitempty"`       // default: 0=Signataire, 1=Observateur
	SignMode         int    `json:"SignMode,omitempty"`         // default: 0=SMS, 1=Email
}

type Transaction struct {
	Name                   string              `json:"Name,omitempty"`
	EndDate                string              `json:"EndDate,omitempty"`  // Format "2020-05-30T00:00:00"
	Reminder               int                 `json:"Reminder,omitempty"` // Number of days between Reminders
	MaximumReminders       int                 `json:"MaximumReminders,omitempty"`
	InvitationEmailSubject string              `json:"InvitationEmailSubject,omitempty"` // default: "Vous êtes invité(e) à signer un document",
	InvitationEmailBody    string              `json:"InvitationEmailBody,omitempty"`    // default: "<p>Bonjour {firstName} {lastName},</p><p><br></p><p>Je vous invite à signer le document &quot;{transactionName}&quot;. Vous pouvez consulter et signer ce document en cliquant sur le bouton ci-dessous :</p><p><br></p><p>{accessLink}</p><p><br></p><p>Vous pourrez signer le document après consultation et vérification de votre identité au moyen d’un code de sécurité.</p>",
	ReminderEmailSubject   string              `json:"ReminderEmailSubject,omitempty"`   // default: "Rappel de signature de document",
	ReminderEmailBody      string              `json:"ReminderEmailBody,omitempty"`      // default: "<p>Bonjour {firstName} {lastName},</p><p><br></p><p>Le document &quot;{transactionName}&quot; est toujours en attente de signature de votre part. Vous pouvez consulter et signer ce document en cliquant sur le bouton ci-dessous :</p><p><br></p><p>{accessLink}</p><p><br></p><p>Pour rappel, après consultation vous pourrez signer ce document au moyen d’un simple code de sécurité.</p>",
	SignatureEmailSubject  string              `json:"SignatureEmailSubject,omitempty"`  // default: "Un utilisateur a signé un document",
	SignatureEmailBody     string              `json:"SignatureEmailBody,omitempty"`     // default: "<p>Bonjour {firstName} {lastName},</p><p><br></p><p>Je vous informe que {signatory.firstName} {signatory.lastName} vient de signer ou valider le document &quot;{transactionName}&quot; que vous avez émis.</p>",
	CompletionEmailSubject string              `json:"CompletionEmailSubject,omitempty"` // default: "Votre document est signé",
	CompletionEmailBody    string              `json:"CompletionEmailBody,omitempty"`    // default: "<p>Bonjour {firstName} {lastName},</p><p><br></p><p>Je vous informe que le document &quot;{transactionName}&quot; a bien été signé.</p>",
	RefusalEmailSubject    string              `json:"RefusalEmailSubject,omitempty"`    // default: "Votre document a été refusé",
	RefusalEmailBody       string              `json:"RefusalEmailBody,omitempty"`       // default: "<p>Bonjour {firstName} {lastName},</p><p><br></p><p>Je vous informe que le document &quot;{transactionName}&quot; a été refusé.</p>",
	Webhook                string              `json:"Webhook,omitempty"`
	IsTest                 bool                `json:"IsTest,omitempty"`
	TransactionMembers     []TransactionMember `json:"TransactionMembers,omitempty"`
}

func respErrorMaker(statusCode int, body io.Reader) (err error) {
	status := "HTTP " + strconv.Itoa(statusCode) + " " + http.StatusText(statusCode)
	if statusCode == 429 {
		err = errors.New(status)
		return
	}
	type errorJSON struct {
		Err    string `json:"error"`
		Errors []struct {
			Err string `json:"error"`
		} `json:"errors"`
	}
	var msg errorJSON
	dec := json.NewDecoder(body)
	err = dec.Decode(&msg)
	if err != nil {
		return err
	}
	var errtxt string
	errtxt += msg.Err
	for i, v := range msg.Errors {
		if i == len(msg.Errors)-1 {
			errtxt += v.Err
		} else {
			errtxt += v.Err + ", "
		}
	}
	if errtxt == "" {
		err = errors.New(status)
	} else {
		err = errors.New(status + ", Message(s): " + errtxt)
	}
	return
}

func (api *API) getResponse(u *url.URL) (ret string, err error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Println("Error in http.NewRequest")
		return
	}
	req.SetBasicAuth(api.Email, api.Key)
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		err = respErrorMaker(resp.StatusCode, resp.Body)
		log.Println("Error because resp.StatusCode: ", resp.StatusCode)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	ret = string(body)
	return
}

func (api *API) postResponse(u *url.URL, contenttype string, payload io.Reader) (id string, err error) {
	req, err := http.NewRequest("POST", u.String(), payload)
	if err != nil {
		log.Println("Error in http.NewRequest")
		return
	}
	if contenttype != "" {
		req.Header.Set("Content-Type", contenttype)
	}
	req.SetBasicAuth(api.Email, api.Key)
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))
	tr := &http.Transport{
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		err = respErrorMaker(resp.StatusCode, resp.Body)
		log.Println("Error because resp.StatusCode: ", resp.StatusCode)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	id = string(body)
	return
}

func (api *API) CreateContact(contact Contact) (id string, err error) {
	u, err := url.Parse(baseURL + "Contacts")
	if err != nil {
		return
	}
	var body bytes.Buffer
	enc := json.NewEncoder(&body)
	if err = enc.Encode(contact); err != nil {
		return
	}
	id, err = api.postResponse(u, "application/json", &body)
	return
}

func (api *API) CreateTransaction(tx Transaction) (id string, err error) {
	u, err := url.Parse(baseURL + "Transactions")
	if err != nil {
		return
	}
	var body bytes.Buffer
	enc := json.NewEncoder(&body)
	if err = enc.Encode(tx); err != nil {
		return
	}
	id, err = api.postResponse(u, "application/json", &body)
	return
}

func (api *API) AddTransactionDocument(txid string, filename string, file io.Reader) (err error) {
	u, err := url.Parse(baseURL + "TransactionFiles")
	if err != nil {
		return
	}
	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)
	_ = writer.WriteField("TransactionId", txid)
	_ = writer.WriteField("FileName", filename)
	part3, errFile3 := writer.CreateFormFile("Document", filename)
	_, errFile3 = io.Copy(part3, file)
	if errFile3 != nil {
		log.Println(errFile3)
		return
	}
	err = writer.Close()
	if err != nil {
		return
	}
	_, err = api.postResponse(u, writer.FormDataContentType(), &payload)
	return
}

func (api *API) AddTransactionMember(member TransactionMember) (id string, err error) {
	u, err := url.Parse(baseURL + "TransactionMembers")
	if err != nil {
		return
	}
	var body bytes.Buffer
	enc := json.NewEncoder(&body)
	if err = enc.Encode(member); err != nil {
		return
	}
	id, err = api.postResponse(u, "application/json", &body)
	return
}

func (api *API) LaunchTransaction(txid string) (err error) {
	u, err := url.Parse(baseURL + "Transactions/LaunchTransaction/" + txid)
	if err != nil {
		return
	}
	_, err = api.postResponse(u, "", nil)
	return
}

func (api *API) GetTransactionStatus(txid string) (status string, err error) {
	u, err := url.Parse(baseURL + "Transactions/Status/" + txid)
	if err != nil {
		return
	}
	status, err = api.getResponse(u)
	return
}

func (api *API) SendTransactionReminders(txid string) (status string, err error) {
	u, err := url.Parse(baseURL + "SendReminders/" + txid)
	if err != nil {
		return
	}
	status, err = api.getResponse(u)
	return
}

func (api *API) AbortTransaction(txid string) (status string, err error) {
	u, err := url.Parse(baseURL + "Abort/" + txid)
	if err != nil {
		return
	}
	status, err = api.getResponse(u)
	return
}

func (api *API) GetAllTransactionDocuments(txid string) (docs string, err error) {
	u, err := url.Parse(baseURL + "Transactions/DownloadDocument/" + txid)
	if err != nil {
		return
	}
	docs, err = api.getResponse(u)
	// TODO : base64 decode
	return
}

func (api *API) GetTransactionProof(txid string) (proof string, err error) {
	u, err := url.Parse(baseURL + "Transactions/DownloadProofFile/" + txid)
	if err != nil {
		return
	}
	proof, err = api.getResponse(u)
	// TODO : base64 decode
	return
}
