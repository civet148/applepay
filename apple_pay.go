package applepay

import (
	"encoding/json"
	"fmt"
	"github.com/civet148/gotools/log"
	"io/ioutil"
	"net/http"
	"strings"
)

/**
 * 21000  App Store无法读取您提供的JSON对象。
 * 21002 该receipt-data属性中的数据格式错误或丢失。
 * 21003 收据无法认证。
 * 21004 您提供的共享密码与您帐户的文件共享密码不匹配。
 * 21005 收据服务器当前不可用。
 * 21006 该收据有效，但订阅已过期。当此状态代码返回到您的服务器时，收据数据也会被解码并作为响应的一部分返回。仅针对自动续订的iOS 6样式交易收据返回。
 * 21007 该收据来自测试环境，但已发送到生产环境以进行验证。而是将其发送到测试环境。
 * 21008 该收据来自生产环境，但是已发送到测试环境以进行验证。而是将其发送到生产环境。
 * 21010 此收据无法授权。就像从未进行过购买一样对待。
 * 21100-21199 内部数据访问错误。
 * 在测试环境中，https://sandbox.itunes.apple.com/verifyReceipt用作URL。在生产中，https://buy.itunes.apple.com/verifyReceipt用作URL。
 */

const (
	APPLE_PAY_VERIFY_URL_PROD    = "https://buy.itunes.apple.com/verifyReceipt"
	APPLE_PAY_VERIFY_URL_SANDBOX = "https://sandbox.itunes.apple.com/verifyReceipt"
)

type ApplePay struct {
	strVerifyUrl     string
	strSharePassword string
	httpClient       *http.Client
}

type requestBody struct {
	ReceiptData string `json:"receipt-data"`
	Password    string `json:"password"`
	//ExcludeOldTransactions bool `json:"exclude-old-transactions"` //Use this field only for app receipts that contain auto-renewable subscriptions
}

type responseBody struct {
	Environment string `json:"environment"`
	IsRetryable bool   `json:"is-retryable"`
	Status      int    `json:"status"`
}

type ApplePayStatus int

const (
	ApplePayStatus_OK                  ApplePayStatus = 0     //收据验证OK
	ApplePayStatus_ErrorJson           ApplePayStatus = 21000 //App Store无法读取您提供的JSON对象。
	ApplePayStatus_ErrorReceiptData    ApplePayStatus = 21002 //该receipt-data属性中的数据格式错误或丢失。
	ApplePayStatus_ErrorReceiptInvalid ApplePayStatus = 21003 //收据无法认证。
	ApplePayStatus_ErrorSharePassword  ApplePayStatus = 21004 //您提供的共享密码与您帐户的文件共享密码不匹配。
	ApplePayStatus_ErrorServer         ApplePayStatus = 21005 //收据服务器当前不可用。
	ApplePayStatus_ErrorReceiptExpired ApplePayStatus = 21006 //该收据有效，但订阅已过期。当此状态代码返回到您的服务器时，收据数据也会被解码并作为响应的一部分返回。仅针对自动续订的iOS 6样式交易收据返回。
	ApplePayStatus_ErrorReceiptSandbox ApplePayStatus = 21007 //该收据来自测试环境，但已发送到生产环境以进行验证。而是将其发送到测试环境。
	ApplePayStatus_ErrorReceiptProd    ApplePayStatus = 21008 //该收据来自生产环境，但是已发送到测试环境以进行验证。而是将其发送到生产环境。
	ApplePayStatus_ErrorReceiptNoAuth  ApplePayStatus = 21010 //此收据无法授权。就像从未进行过购买一样对待。
	ApplePayStatus_ErrorInternalMin    ApplePayStatus = 21100 //21100~21199 内部数据访问错误。
	ApplePayStatus_ErrorInternalMax    ApplePayStatus = 21199 //21100~21199 内部数据访问错误。
)

func (s ApplePayStatus) String() string {

	switch s {
	case ApplePayStatus_OK:
		return "ApplePayStatus_OK"
	case ApplePayStatus_ErrorJson:
		return "ApplePayStatus_ErrorJson"
	case ApplePayStatus_ErrorReceiptData:
		return "ApplePayStatus_ErrorReceiptData"
	case ApplePayStatus_ErrorReceiptInvalid:
		return "ApplePayStatus_ErrorReceiptInvalid"
	case ApplePayStatus_ErrorSharePassword:
		return "ApplePayStatus_ErrorSharePassword"
	case ApplePayStatus_ErrorServer:
		return "ApplePayStatus_ErrorServer"
	case ApplePayStatus_ErrorReceiptExpired:
		return "ApplePayStatus_ErrorReceiptExpired"
	case ApplePayStatus_ErrorReceiptSandbox:
		return "ApplePayStatus_ErrorReceiptSandbox"
	case ApplePayStatus_ErrorReceiptProd:
		return "ApplePayStatus_ErrorReceiptProd"
	case ApplePayStatus_ErrorReceiptNoAuth:
		return "ApplePayStatus_ErrorReceiptNoAuth"
	default:
		if s >= ApplePayStatus_ErrorInternalMin && s <= ApplePayStatus_ErrorInternalMax {
			return fmt.Sprintf("ApplePayStatus_ErrorInternal<%d>", s)
		}
	}
	return "ApplePayStatus_Unknown"
}

func (s ApplePayStatus) GoString() string {
	return s.String()
}

func NewApplePay(strSharePassword, strVerifyUrl string) *ApplePay {

	if strVerifyUrl == "" {
		log.Errorf("NewApplePay verify url is nil")
		return nil
	}
	return &ApplePay{
		strVerifyUrl:     strVerifyUrl,
		strSharePassword: strSharePassword,
		httpClient:       &http.Client{},
	}
}

//校验苹果支付回执收据
func (a *ApplePay) VerifyReceipt(strReceipt string) (ok bool, err error) {

	var response *responseBody

	log.Infof("VerifyReceipt receipt [%s]", strReceipt)

	if response, err = a.postVerifyRequest(&requestBody{
		ReceiptData: strReceipt,
		Password:    a.strSharePassword,
	}); err != nil {
		log.Errorf("VerifyReceipt post verify request error [%s]", err)
		return
	}

	var status = ApplePayStatus(response.Status)
	if status == ApplePayStatus_OK {
		ok = true
	} else {
		err = fmt.Errorf("VerifyReceipt got verify response status [%s] => [%d]", status, status)
		log.Errorf(err.Error())
	}

	log.Infof("VerifyReceipt receipt [%s] ok [%v]", strReceipt, ok)
	return
}

func (a *ApplePay) postVerifyRequest(request *requestBody) (response *responseBody, err error) {

	var data []byte

	data, err = json.Marshal(request)
	if err != nil {
		log.Errorf("VerifyReceipt json.Marshal(%+v) error [%v]", request, err.Error())
		return
	}

	var httpRequest *http.Request
	if httpRequest, err = http.NewRequest(http.MethodPost, a.strVerifyUrl, strings.NewReader(string(data))); err != nil {
		log.Errorf("VerifyReceipt http.NewRequest return error [%v]", err.Error())
		return
	}

	var httpResponse *http.Response
	httpRequest.Header.Set("Content-Type", "application/json") //内容类型(JSON)

	if httpResponse, err = a.httpClient.Do(httpRequest); err != nil {
		log.Errorf("VerifyReceipt %v to [%v] with data [%s] failed [%v]", http.MethodPost, a.strVerifyUrl, data, err.Error())
		return
	}
	defer httpResponse.Body.Close()

	if httpResponse == nil {
		err = fmt.Errorf("VerifyReceipt %v to [%v] with data [%s] got nil response", http.MethodPost, a.strVerifyUrl, data)
		log.Error(err.Error())
		return
	}

	if httpResponse.StatusCode != http.StatusOK {
		err = fmt.Errorf("VerifyReceipt %v to [%v] with data [%s] got response status [%d]", http.MethodPost, a.strVerifyUrl, data, httpResponse.StatusCode)
		log.Error(err.Error())
		return
	}

	var body []byte
	if body, err = ioutil.ReadAll(httpResponse.Body); err != nil {
		log.Errorf("VerifyReceipt ioutil.ReadAll from http response body failed [%v]", err.Error())
		return
	}

	log.Debugf("VerifyReceipt http response [%s]", body)

	response = &responseBody{}
	if err = json.Unmarshal(body, response); err != nil {
		log.Errorf("VerifyReceipt json.Unmarshal from http response body failed [%v]", err.Error())
		return nil, err
	}
	return
}
