package gateway

import (
	"sync"
	"net/url"
	"strings"
	"bytes"
	"net/http"
	"strconv"
	"io/ioutil"
)

const (
	TRANSACTION_TYPE_SALE  = "D"
	TRANSACTION_TYPE_PREAUTH  = "P"
	TRANSACTION_TYPE_SETTLE  = "S"
	TRANSACTION_TYPE_REAUTH = "J"
	TRANSACTION_TYPE_OFFLINE  = "O"
	TRANSACTION_TYPE_VOID  = "V"
	TRANSACTION_TYPE_CREDIT  = "C"
	TRANSACTION_TYPE_REFUND  = "U"
	TRANSACTION_TYPE_VERIFY = "A"
	TRANSACTION_TYPE_TOKENIZE  = "T"
	TRANSACTION_TYPE_DETOKENIZE  = "X"
	TRANSACTION_TYPE_BATCHCLOSE  = "Z"
	GATEWAY_URL_CERT = "https://cert.merchante-solutions.com/mes-api/tridentApi"
	GATEWAY_URL_TEST = "https://test.merchante-solutions.com/mes-api/tridentApi"
	GATEWAY_URL_LIVE = "https://api.merchante-solutions.com/mes-api/tridentApi"
)

type Transaction struct{
	requestParameters []map[string]string
	hostUrl string
	sync.RWMutex
}

type response struct{
	responseList map[string]string
	sync.RWMutex
}

type libHTTP struct{
	requestUrl string
	requestString string
	rawResponse string
	sync.RWMutex
}

func (requestGateway *Transaction) Init(gatewayUrl string, transactionType string) () {
	requestGateway.Lock()
	requestGateway.requestParameters = make([]map[string]string, 0)
	requestGateway.Unlock()

	requestGateway.AddParameter("transaction_type", transactionType)
	requestGateway.Lock()
	requestGateway.hostUrl = gatewayUrl
	requestGateway.Unlock()
}

func (requestGateway *Transaction) AddParameter(key string, value string) () {
	requestGateway.Lock()
	requestGateway.requestParameters = append(requestGateway.requestParameters, map[string]string{
		key: value,
	})
	requestGateway.Unlock()
}

func (requestGateway *Transaction) RequestString() (request string) {
	requestGateway.RLock()
	for _, keyValuePair := range requestGateway.requestParameters {
		for key, value := range keyValuePair {
			request += key + "=" + url.QueryEscape(value) + "&"
		}
	}
	requestGateway.RUnlock()
	return request[:len(request)-1]
}


func (requestGateway *Transaction) HostUrl(gatewayUrl string) () {
	requestGateway.Lock()
	requestGateway.hostUrl = gatewayUrl
	requestGateway.Unlock()
}

func (requestGateway *Transaction) AddCredentials(profileId string, profileKey string) () {
	requestGateway.AddParameter("profile_id", profileId)
	requestGateway.AddParameter("profile_key", profileKey)
}

func (requestGateway *Transaction) AddCardData(cardNum string, expDate string) () {
	requestGateway.AddParameter("card_number", cardNum)
	requestGateway.AddParameter("card_exp_date", expDate)
}

func (requestGateway *Transaction) AddTokenData(token string, expDate string) () {
	requestGateway.AddParameter("card_id", token)
	requestGateway.AddParameter("card_exp_date", expDate)
}

func (requestGateway *Transaction) AddAVSData(address string, zipCode string) () {
	requestGateway.AddParameter("cardholder_street_address", address)
	requestGateway.AddParameter("cardholder_zip", zipCode)
}

func (requestGateway *Transaction) AddInvoice(invoice string) () {
	requestGateway.AddParameter("invoice_number", invoice)
}

func (requestGateway *Transaction) AddClientRef(ref string) () {
	requestGateway.AddParameter("client_reference_number", ref)
}

func (requestGateway *Transaction) AddAmount(amount string) () {
	requestGateway.AddParameter("transaction_amount", amount)
}

func (requestGateway *Transaction) AddTranId(tranId string) () {
	requestGateway.AddParameter("transaction_id", tranId)
}

func (requestGateway *Transaction) Run() (*response, error) {
	requestGateway.RLock()
	host := requestGateway.hostUrl
	postURL := requestGateway.RequestString()
	requestGateway.RUnlock()
	httpInstance := libHTTP{}
	httpInstance.Init(host, postURL)
	parsedResponse := response{}
	rawResponse, err := httpInstance.Run()
	if err != nil {
		return &parsedResponse, err
	}
	parsedResponse.Init(rawResponse)
	return &parsedResponse, err
}

func (httpSender *libHTTP) Init(urlString string, requestString string) () {
	httpSender.Lock()
	httpSender.requestUrl = urlString
	httpSender.requestString = requestString
	httpSender.Unlock()
}

func (httpSender *libHTTP) Run() (string, error) {
	httpSender.RLock()
	apiUrl := httpSender.requestUrl
	req := httpSender.requestString
	httpSender.RUnlock()
	u, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		return "", err
	}
	client := &http.Client{}
	r, err := http.NewRequest("POST", u.String(), bytes.NewBufferString(req))
	if err != nil {
		return "", err
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(req)))


	responseData, err := client.Do(r)

	defer r.Body.Close()
	if responseData.StatusCode == 200 {
		bodyBytes, err := ioutil.ReadAll(responseData.Body)
		if err != nil {
			return "", err
		}
		return string(bodyBytes), err
	}
	return "", err
}

func (gatewayResponse *response) Init(responseString string) (*response) {
	gatewayResponse.Lock()
	gatewayResponse.responseList = make(map[string]string, 0)
	pairs := strings.Split(responseString, "&")
	for _, pair := range pairs {
		npv := strings.Split(pair, "=")
		gatewayResponse.responseList[npv[0]] = npv[1]
	}
	gatewayResponse.Unlock()
	return gatewayResponse
}

func (gatewayResponse *response) GetValue(key string) (value string) {
	gatewayResponse.RLock()
	val, ok := gatewayResponse.responseList[key]
	gatewayResponse.RUnlock()
	if ok {
		return val
	}
	return ""
}

func (gatewayResponse *response) GetRespText() (string) {
	return gatewayResponse.GetValue("auth_response_text")
}

func (gatewayResponse *response) GetTranId() (string) {
	return gatewayResponse.GetValue("transaction_id")
}

func (gatewayResponse *response) GetErrorCode() (string) {
	return gatewayResponse.GetValue("error_code")
}

func (gatewayResponse *response) GetAvsResult() (string) {
	return gatewayResponse.GetValue("avs_result")
}

func (gatewayResponse *response) GetCvvResult() (string) {
	return gatewayResponse.GetValue("cvv2_result")
}

func (gatewayResponse *response) GetAuthCode() (string) {
	return gatewayResponse.GetValue("auth_code")
}

func (gatewayResponse *response) IsApproved() (bool) {
	code := gatewayResponse.GetErrorCode()
	return code == "000" || code == "085"
}
