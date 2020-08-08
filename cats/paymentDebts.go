package cats

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

type debtsResponse struct {
	SEcho                string
	ITotalRecords        int
	ITotalDisplayRecords int
	AaData               [][]string
	AsWarnings           string
	LimitedAccess        bool
}

func (r *debtsResponse) ToJSON() (string, error) {
	type realTime struct {
		Kodi                string
		Periudha            string
		KodiTeArdhurave     string
		PergjegjesiTatimore string
		LlojiGjobes         string
		DetyrimPapaguar     string
		GjobePapaguar       string
		InteresPapaguar     string
		TotaliPapaguar      string
		Numer               string
	}
	RealTime := []realTime{}
	for _, a := range r.AaData {
		if len(a) < 9 { //ndryshe do kemi problem me poshte kur te kapim elementet
			return "", fmt.Errorf("debtsResponse ToJSON array id tabeles me pak se 9 colona: %v", a)
		}
		RealTime = append(RealTime, realTime{
			Kodi:                a[0],
			Periudha:            a[1],
			KodiTeArdhurave:     a[2],
			PergjegjesiTatimore: a[3],
			LlojiGjobes:         a[4],
			DetyrimPapaguar:     a[5],
			GjobePapaguar:       a[6],
			InteresPapaguar:     a[7],
			TotaliPapaguar:      a[8],
			Numer:               a[9],
		})
	}
	b, err := json.Marshal(RealTime)
	if err != nil {
		return "", fmt.Errorf("debtsResponse ToJSON problem ne konvertim JSON")
	}
	return string(b), nil
}

func (a *accountingFields) debitsPaymentPostData() map[string]string {
	return map[string]string{
		"sEcho":          "1",
		"iColumns":       "10",
		"sColumns":       ",,,,,,,,,",
		"iDisplayStart":  "0",
		"iDisplayLength": "10",
		"mDataProp_0":    "0",
		"bSortable_0":    "false",
		"mDataProp_1":    "1",
		"bSortable_1":    "true",
		"mDataProp_2":    "2",
		"bSortable_2":    "true",
		"mDataProp_3":    "3",
		"bSortable_3":    "true",
		"mDataProp_4":    "4",
		"bSortable_4":    "true",
		"mDataProp_5":    "5",
		"bSortable_5":    "false",
		"mDataProp_6":    "6",
		"bSortable_6":    "false",
		"mDataProp_7":    "7",
		"bSortable_7":    "false",
		"mDataProp_8":    "8",
		"bSortable_8":    "false",
		"mDataProp_9":    "9",
		"bSortable_9":    "true",
		"iSortCol_0":     "1",
		"sSortDir_0":     "desc",
		"iSortingCols":   "1",
		"Selection":      "",
		"InterestDate":   time.Now().Format("02.01.2006"),
		"TaxpayerId":     "",
		"TaxPeriodId":    "-1",
	}
}

func (a *accountingFields) debitsPaymentDownloadPostData(PeriodId string, RevenueCode string, tokenName string, tokenVal string) map[string]string {
	return map[string]string{
		"PeriodId":     PeriodId,
		"RevenueCode":  RevenueCode,
		"InterestDate": time.Now().Format("02.01.2006"),
		"tokenName":    tokenVal,
	}
}

func (c *CatsUserData) debtPayment() error {
	collector := c.collector.Clone()
	var accountDataError string
	c.collector.OnHTML("form", func(e *colly.HTMLElement) {
		if (e.Attr("method") == "post") && (e.ChildAttr("input", "name") == "__RequestVerificationToken") {
			c.formToken.name = e.ChildAttr("input", "name")
			c.formToken.value = e.ChildAttr("input", "value")
		}
	})
	collector.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			accountDataError = "error getting debtPayment status != 200"
			log.Warningln(accountDataError)
			log.Warningln(r.Request.URL)
			log.Warningln(r.Request.Marshal())
			log.Warningln(r.StatusCode)
			log.Warningln(r.Headers)
			log.Warningln(r.Body)
		} else {
			token, er := getAjaxToken(r.Body)
			if er != nil {
				accountDataError = "debtPayment does not have ajax request token" + er.Error()
				log.Warningln(accountDataError)
				log.Warningln(r.Request.URL)
				log.Warningln(r.Request.Marshal())
				log.Warningln(r.StatusCode)
				log.Warningln(r.Headers)
				log.Warningln(string(r.Body))
			} else {
				c.ajaxToken = token
			}
			token, er = getAntiForgery(r.Body)
			if er != nil {
				accountDataError = "debtPayment does not have anti forgery request token" + er.Error()
				log.Warningln(accountDataError)
				log.Warningln(r.Request.URL)
				log.Warningln(r.Request.Marshal())
				log.Warningln(r.StatusCode)
				log.Warningln(r.Headers)
				log.Warningln(string(r.Body))
			} else {
				c.antiForgeryToken = *token
			}
			tableTokes, er := getTableToken(r.Body)
			if (c.accountingDebtsPaymentHasAllFields()) && (er != nil) {
				accountDataError = er.Error()
			} else {
				for key, val := range tableTokes {
					c.tableTokens = append(c.tableTokens, &tokenFromHTML{
						name:  key,
						value: val,
					})
				}
			}
		}
	})
	collector.OnScraped(func(r *colly.Response) {
		//fmt.Println("Finished", r.Request.URL)
	})
	collector.OnError(func(r *colly.Response, err error) {
		accountDataError = "debtPayment error Error: " + r.Request.URL.String() + err.Error()
		log.Warningln(accountDataError)
		log.Warningln(r.Request.URL)
		log.Warningln(r.Request.Marshal())
		log.Warningln(r.StatusCode)
		log.Warningln(r.Headers)
		log.Warningln(string(r.Body))
	})
	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	collector.RedirectHandler = redirHandler

	//TODO use timeout
	collector.UserAgent = userAgent
	err := collector.Visit(efilingURL + accountingPaymentDebts)
	if err != nil {
		return err
	} else if accountDataError != "" {
		return errors.New(accountDataError)
	}
	c.collector = collector.Clone()
	return nil
}

func (c *CatsUserData) AccountingDebtPayment() error {
	err := c.debtPayment()
	if err != nil {
		return err
	}
	collector := c.collector.Clone()

	var replyError string
	collector.OnResponse(func(r *colly.Response) {
		if (r.StatusCode != 200) || (r.Headers.Get("Content-Type") != "application/json; charset=utf-8") {
			replyError = "AccountingDebtPayment status != 200 or not json"
		} else {
			e := json.Unmarshal(r.Body, c.AccountingFields.DebtsPayment)
			if e != nil {
				replyError = "AccountingDebtPayment wrong data json " + e.Error()
				log.Warningln(replyError)
				log.Warningln(r.Request.URL)
				log.Warningln(r.Request.Marshal())
				log.Warningln(r.StatusCode)
				log.Warningln(r.Headers)
				log.Warningln(string(r.Body))
			}
			_, e = c.AccountingFields.DebtsPayment.ToJSON()
			if e != nil {
				replyError = e.Error()
				log.Warningln(replyError)
				log.Warningln(r.Request.URL)
				log.Warningln(r.Request.Marshal())
				log.Warningln(r.StatusCode)
				log.Warningln(r.Headers)
				log.Warningln(string(r.Body))
			}
		}
	})
	collector.OnScraped(func(r *colly.Response) {
		//fmt.Println("Finished", r.Request.URL)
	})
	collector.OnError(func(r *colly.Response, err error) {
		replyError = "Error: " + r.Request.URL.String() + err.Error()
		log.Warningln(replyError)
		log.Warningln(r.Request.URL)
		log.Warningln(r.Request.Marshal())
		log.Warningln(r.StatusCode)
		log.Warningln(r.Headers)
		log.Warningln(string(r.Body))
	})
	collector.OnRequest(func(r *colly.Request) {
		for _, val := range c.tableTokens {
			r.Headers.Set(val.name, val.value)
		}
		r.Headers.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		r.Headers.Set("X-Requested-With", "XMLHttpRequest")
		r.Headers.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		r.Headers.Set("Origin", efilingURL)
		r.Headers.Set("Referer", efilingURL+accountingPaymentDebts)

		fmt.Println("Visiting", r.URL)
	})
	collector.RedirectHandler = redirHandler
	//TODO use timeout
	collector.UserAgent = userAgent
	err = collector.Post(efilingURL+accountingPaymentGetDebts, c.AccountingFields.debitsPaymentPostData())
	if err != nil {
		return err
	} else if replyError != "" {
		return errors.New(replyError)
	}

	c.collector = collector.Clone()
	return nil
}

func (c *CatsUserData) AccountingDebtPaymentOrderDownload(PeriodId string, RevenueCode string) error {
	collector := c.collector.Clone()
	var replyError string
	collector.OnResponse(func(r *colly.Response) {
		if (r.StatusCode != 200) || (r.Headers.Get("Content-Type") != "application/pdf") {
			replyError = "AccountingDebtPaymentOrderDownload status != 200 or not pdf"
			log.Warningln(replyError)
			log.Warningln(r.Request.URL)
			log.Warningln(r.Request.Marshal())
			log.Warningln(r.StatusCode)
			log.Warningln(r.Headers)
			log.Warningln(r.Body)
		} else {
			fileName := r.FileName()
			err := r.Save(fileName)
			if err != nil {
				replyError = err.Error()
				log.Warningln(replyError)
				log.Warningln(r.Request.URL)
				log.Warningln(r.Request.Marshal())
				log.Warningln(r.StatusCode)
				log.Warningln(r.Headers)
				log.Warningln(string(r.Body))
			}
		}
	})
	collector.OnScraped(func(r *colly.Response) {
		//fmt.Println("Finished", r.Request.URL)
	})
	collector.OnError(func(r *colly.Response, err error) {
		replyError = "Error: " + r.Request.URL.String() + err.Error()
		log.Warningln(replyError)
		log.Warningln(r.Request.URL)
		log.Warningln(r.Request.Marshal())
		log.Warningln(r.StatusCode)
		log.Warningln(r.Headers)
		log.Warningln(string(r.Body))
	})
	collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set(c.antiForgeryToken.name, c.antiForgeryToken.value)
		r.Headers.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		r.Headers.Set("X-Requested-With", "XMLHttpRequest")
		r.Headers.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		r.Headers.Set("Origin", efilingURL)
		r.Headers.Set("Referer", efilingURL+accountingPaymentDebts)

		fmt.Println("Visiting", r.URL)
	})
	collector.RedirectHandler = redirHandler
	//TODO use timeout
	collector.UserAgent = userAgent
	err := collector.Post(efilingURL+accountingPaymentDebtsDownload,
		c.AccountingFields.debitsPaymentDownloadPostData(PeriodId, RevenueCode, c.antiForgeryToken.name, c.antiForgeryToken.value))
	if err != nil {
		return err
	} else if replyError != "" {
		return errors.New(replyError)
	}

	c.collector = collector.Clone()
	return nil
}
