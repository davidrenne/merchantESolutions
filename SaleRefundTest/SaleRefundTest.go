package main

import (
	"github.com/davidrenne/merchantESolutions"
	"fmt"
)

func main() {
	g := gateway.Transaction{}
	g.Init(gateway.GATEWAY_URL_CERT, gateway.TRANSACTION_TYPE_SALE)
	g.AddCredentials("profileid", "profilekey")
	g.AddCardData("4012888812348882", "1216")
	g.AddAVSData("123 N. Main", "55555")
	g.AddAmount("1.00")
	salesResp, err := g.Run()

	if err == nil {
		fmt.Println("Sale error code: " + salesResp.GetErrorCode())
		fmt.Println("Sale resp text: " + salesResp.GetRespText())
		g := gateway.Transaction{}
		g.Init(gateway.GATEWAY_URL_CERT, gateway.TRANSACTION_TYPE_SALE)
		g.AddCredentials("profileid", "profilekey")
		g.AddTranId(salesResp.GetTranId())
		refundResp, err := g.Run()
		if err == nil {
			fmt.Println("Sale error code: "+ refundResp.GetErrorCode())
			fmt.Println("Sale resp text: "+ refundResp.GetRespText())
		} else {
			fmt.Println("Error:", err)
		}
	} else {
		fmt.Println("Error:", err)
	}
}