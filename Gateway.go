package gateway

import (
	"sync"
	"net/url"
	"strings"
	"bytes"
	"net/http"
	"strconv"
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

type gateway struct{
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
	sync.RWMutex
}

var Request = gateway{}

func (requestGateway *gateway) Init(gatewayUrl string, transactionType string) () {
	requestGateway.Lock()
	requestGateway.requestParameters = make([]map[string]string, 0)
	requestGateway.AddParameter("transaction_type", transactionType)
	requestGateway.hostUrl = gatewayUrl
	requestGateway.Unlock()
}

func (requestGateway *gateway) AddParameter(key string, value string) () {
	requestGateway.Lock()
	requestGateway.requestParameters = append(requestGateway.requestParameters, map[string]string{
		key: value,
	})
	requestGateway.Unlock()
}

func (requestGateway *gateway) RequestString() (request string) {
	requestGateway.RLock()
	for keyValuePair := range requestGateway.requestParameters {
		for key, value := range keyValuePair {
			request += key + "=" + url.QueryEscape(value) + "&"
		}
	}
	requestGateway.RUnlock()
	return request[:-1]
}


func (requestGateway *gateway) HostUrl(gatewayUrl string) () {
	requestGateway.Lock()
	requestGateway.hostUrl = gatewayUrl
	requestGateway.Unlock()
}

func (requestGateway *gateway) AddCredentials(profileId string, profileKey string) () {
	requestGateway.AddParameter("profile_id", profileId)
	requestGateway.AddParameter("profile_key", profileKey)
}

func (requestGateway *gateway) AddCardData(cardNum string, expDate string) () {
	requestGateway.AddParameter("card_number", cardNum)
	requestGateway.AddParameter("card_exp_date", expDate)
}

func (requestGateway *gateway) AddTokenData(token string, expDate string) () {
	requestGateway.AddParameter("card_id", token)
	requestGateway.AddParameter("card_exp_date", expDate)
}

func (requestGateway *gateway) AddAVSData(address string, zipCode string) () {
	requestGateway.AddParameter("cardholder_street_address", address)
	requestGateway.AddParameter("cardholder_zip", zipCode)
}

func (requestGateway *gateway) AddInvoice(invoice string) () {
	requestGateway.AddParameter("invoice_number", invoice)
}

func (requestGateway *gateway) AddClientRef(ref string) () {
	requestGateway.AddParameter("client_reference_number", ref)
}

func (requestGateway *gateway) AddAmount(amount string) () {
	requestGateway.AddParameter("transaction_amount", amount)
}

func (requestGateway *gateway) AddTranId(tranId string) () {
	requestGateway.AddParameter("transaction_id", tranId)
}

func (requestGateway *gateway) Run() (parsedResponse response, err error) {
	requestGateway.RLock()
	host := requestGateway.hostUrl
	postURL := requestGateway.RequestString()
	requestGateway.RUnlock()
	httpInstance := libHTTP{}
	httpInstance.Init(host, postURL)
	response, err := httpInstance.Run()
	if err == nil {
		return
	}
	parsedResponse = response{}
	return parsedResponse.Init(response.Body)
}

func (httpSender *libHTTP) Init(urlString string, requestString string) () {
	httpSender.Lock()
	httpSender.requestUrl = urlString
	httpSender.requestString = requestString
	httpSender.Unlock()
}

func (httpSender *libHTTP) Run() (*http.Response, error) {
	httpSender.RLock()
	apiUrl := httpSender.requestUrl
	req := httpSender.requestString
	httpSender.RUnlock()
	u, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		return
	}
	client := &http.Client{}
	r, err := http.NewRequest("POST", u.String(), bytes.NewBufferString(req))
	if err != nil {
		return
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(req)))
	return client.Do(r)
}

func ( *response) Init(responseString string) (gatewayResponse *response) {
	gatewayResponse.Lock()
	gatewayResponse.responseList = make(map[string]string, 0)
	pairs := strings.Split(responseString, "&")
	for pair := range pairs {
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
