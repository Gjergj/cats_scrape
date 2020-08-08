package cats

import (
	"bytes"
	"encoding/json"
	"errors"
)

const (
	efilingURL                     string = "https://efiling.tatime.gov.al"
	loginPath                      string = efilingURL + "/"
	accountingInterestPath         string = "/cats_public/Accounting/Interest/Index"
	accountingRealTimeDebits       string = "/cats_public/Accounting/Interest/AccountingRealTimeDebitsByAjax"
	accountingPaymentDebts         string = "/cats_public/Accounting/PaymentOrder/Debts"
	accountingPaymentGetDebts      string = "/cats_public/Accounting/PaymentOrder/GetDebts"
	accountingPaymentDebtsDownload string = "/cats_public/Accounting/PaymentOrder/Print"
	isAuthPath                     string = "/cats_public/Account/IsAuthenticated?_="
	userAgent                      string = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"
)

type tokenFromHTML struct {
	name  string
	value string
}

func getTableToken(html []byte) (map[string]string, error) {
	// get ajax token
	tokenBeginSearch := `"headers": `

	beginOfJSON := bytes.Index(html, []byte(tokenBeginSearch))
	if beginOfJSON < 1 {
		return nil, errors.New("error getting table json from html")
	}
	beginOfJSON = beginOfJSON + len(tokenBeginSearch)

	tokenEndSearch := string(`' },`)
	endOfJSON := bytes.Index(html, []byte(tokenEndSearch))
	if endOfJSON < 1 {
		return nil, errors.New("error table json from html")
	}
	endOfJSON += 3 //te perfshije edhe fundin e json, ' }

	jsonByteArray := html[beginOfJSON:endOfJSON]
	jsonByteArray = bytes.ReplaceAll(jsonByteArray, []byte("'"), []byte("\""))
	tableTokens := make(map[string]string)
	err := json.Unmarshal(jsonByteArray, &tableTokens)
	if err != nil {
		return nil, errors.New("AccountingInterest tokens, error parsing json " + err.Error())
	}
	return tableTokens, nil
}

func getAjaxToken(html []byte) (*tokenFromHTML, error) {
	tokens := tokenFromHTML{}
	// get ajax token
	tokenBeginSearch := "xhr.setRequestHeader('"

	beginTokenName := bytes.Index(html, []byte(tokenBeginSearch))
	if beginTokenName < 1 {
		return nil, errors.New("GetAjaxToken error getting __RequestVerificationToken from html")
	}
	beginTokenName += bytes.Index(html[beginTokenName:], []byte("'")) + 1

	endTokenName := bytes.Index(html[beginTokenName:], []byte("'"))
	if (beginTokenName < 1) && (beginTokenName > beginTokenName+len(tokenBeginSearch)) {
		return nil, errors.New("error getting __RequestVerificationToken from html")
	}
	reqTokenName := html[beginTokenName : beginTokenName+endTokenName]

	tokenValBeginSearch := `xhr.setRequestHeader('__RequestVerificationToken', '`
	beginTokenVal := bytes.Index(html, []byte(tokenValBeginSearch)) + len(tokenValBeginSearch)
	if (beginTokenVal < 1) && (beginTokenVal > beginTokenVal+len(tokenValBeginSearch)) {
		return nil, errors.New("error getting __RequestVerificationToken from html")
	}
	endTokenVal := bytes.Index(html[beginTokenVal:], []byte("');")) + beginTokenVal
	if (endTokenVal < 1) && (endTokenVal < beginTokenVal) {
		return nil, errors.New("error getting __RequestVerificationToken from html")
	}
	reqTokenVal := html[beginTokenVal:endTokenVal]

	tokens.name = string(reqTokenName)
	tokens.value = string(reqTokenVal)
	return &tokens, nil
}
func getAntiForgery(html []byte) (*tokenFromHTML, error) {
	tokens := tokenFromHTML{}
	tokenBeginSearch := `downloader.settings.AntiForgery = "`

	beginTokenName := bytes.Index(html, []byte(tokenBeginSearch)) + len(tokenBeginSearch)
	//beginTokenName += bytes.Index(html[beginTokenName:], []byte("'")) + 1
	if beginTokenName < 1 {
		return nil, errors.New("error downloader.settings.AntiForgery from html")
	}

	endTokenName := bytes.Index(html[beginTokenName:], []byte(`";`))
	if (beginTokenName < 1) && (beginTokenName > beginTokenName+len(tokenBeginSearch)) {
		return nil, errors.New("error getting __RequestVerificationToken from html")
	}
	reqTokenName := html[beginTokenName : beginTokenName+endTokenName]
	tokens.name = "__RequestVerificationToken"
	tokens.value = string(reqTokenName)
	return &tokens, nil
}
