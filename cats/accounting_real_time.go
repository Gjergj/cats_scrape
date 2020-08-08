package cats

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

type accountingFields struct {
	RealTimeDebitsResponse *realTimeDebitsResponse
	DebtsPayment           *debtsResponse
}
type realTimeDebitsResponse struct {
	SEcho                string
	ITotalRecords        int
	ITotalDisplayRecords int
	AaData               [][]string
	AsWarnings           string
	LimitedAccess        bool
}

func (r *realTimeDebitsResponse) ToJSON() (string, error) {
	type realTime struct {
		PergjegjesiTatimore   string
		DetyrimiKryesor       string
		GjobaRegjistruara     string
		GjobaPaRegjistruara   string
		InteresiRegjistruar   string
		InteresiPaRegjistruar string
		Totali                string
	}
	RealTime := []realTime{}
	for _, a := range r.AaData {
		if len(a) < 7 { //ndryshe do kemi problem me poshte kur te kapim elementet
			return "", fmt.Errorf("realTimeDebits ToJSON array id tabeles me pake se 7 colona: %v", a)
		}
		RealTime = append(RealTime, realTime{
			PergjegjesiTatimore:   a[0],
			DetyrimiKryesor:       a[1],
			GjobaRegjistruara:     a[2],
			GjobaPaRegjistruara:   a[3],
			InteresiRegjistruar:   a[4],
			InteresiPaRegjistruar: a[5],
			Totali:                a[6],
		})
	}
	b, err := json.Marshal(RealTime)
	if err != nil {
		return "", fmt.Errorf("realTimeDebits ToJSON problem ne konvertim JSON")
	}
	return string(b), nil
}

func (a *accountingFields) RealTimeDebitsPostData() map[string]string {
	return map[string]string{
		"sEcho":           "1",
		"iColumns":        "7",
		"sColumns":        ",,,,,,",
		"iDisplayStart":   "0",
		"iDisplayLength":  "10",
		"mDataProp_0":     "0",
		"bSortable_0":     "true",
		"mDataProp_1":     "1",
		"bSortable_1":     "true",
		"mDataProp_2":     "2",
		"bSortable_2":     "true",
		"mDataProp_3":     "3",
		"bSortable_3":     "true",
		"mDataProp_4":     "4",
		"bSortable_4":     "true",
		"mDataProp_5":     "5",
		"bSortable_5":     "true",
		"mDataProp_6":     "6",
		"bSortable_6":     "true",
		"iSortCol_0":      "0",
		"sSortDir_0":      "asc",
		"iSortingCols":    "1",
		"Selection":       "",
		"DebitsUntilDate": time.Now().Format("02.01.2006"),
		"TaxpayerId":      "", //todo cuditerisht kthen pergjigje edhe me bosh, 331841
	}
}

func (c *CatsUserData) accountingInterest() error {
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
			accountDataError = "error getting AccountingInterest status != 200"
			log.Warningln(accountDataError)
			log.Warningln(r.Request.URL)
			log.Warningln(r.Request.Marshal())
			log.Warningln(r.StatusCode)
			log.Warningln(r.Headers)
			log.Warningln(string(r.Body))
		} else {
			token, er := getAjaxToken(r.Body)
			if er != nil {
				accountDataError = "accountingInterest does not have ajax request token" + er.Error()
				log.Warningln(accountDataError)
				log.Warningln(r.Request.URL)
				log.Warningln(r.Request.Marshal())
				log.Warningln(r.StatusCode)
				log.Warningln(r.Headers)
				log.Warningln(string(r.Body))
			} else {
				c.ajaxToken = token
			}
			tableTokes, er := getTableToken(r.Body)
			if (c.accountingRealTimeDebitsHasAllFields()) && (er != nil) {
				accountDataError = er.Error()
				log.Warningln(accountDataError)
				log.Warningln(r.Request.URL)
				log.Warningln((r.Request.Marshal()))
				log.Warningln(r.StatusCode)
				log.Warningln(r.Headers)
				log.Warningln(string(r.Body))
			} else {
				for key, val := range tableTokes {
					c.tableTokens = append(c.tableTokens, &tokenFromHTML{
						name:  key,
						value: val,
					})
				}
			}
			fmt.Println(c.AccountingFields)
		}
	})
	collector.OnScraped(func(r *colly.Response) {
		//fmt.Println("Finished", r.Request.URL)
	})
	collector.OnError(func(r *colly.Response, err error) {
		println("Error: ", r.Request.URL, err.Error())
	})
	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	collector.RedirectHandler = redirHandler

	//TODO use timeout
	collector.UserAgent = userAgent
	err := collector.Visit(efilingURL + accountingInterestPath)
	if err != nil {
		return err
	} else if accountDataError != "" {
		return errors.New(accountDataError)
	}
	c.collector = collector.Clone()
	return nil
}

func (c *CatsUserData) AccountingRealTimeDebits() error {
	err := c.accountingInterest()
	if err != nil {
		return err
	}
	collector := c.collector.Clone()

	var replyError string
	collector.OnResponse(func(r *colly.Response) {
		if (r.StatusCode != 200) || (r.Headers.Get("Content-Type") != "application/json; charset=utf-8") {
			replyError = "AccountingRealTimeDebits status != 200 or not json"
			log.Warningln(replyError)
			log.Warningln(r.Request.URL)
			log.Warningln(r.Request.Marshal())
			log.Warningln(r.StatusCode)
			log.Warningln(r.Headers)
			log.Warningln(string(r.Body))
		} else {
			e := json.Unmarshal(r.Body, c.AccountingFields.RealTimeDebitsResponse)
			if e != nil {
				replyError = "AccountingRealTimeDebits wrong data json " + e.Error()
				log.Warningln(replyError)
				log.Warningln(r.Request.URL)
				log.Warningln(r.Request.Marshal())
				log.Warningln(r.StatusCode)
				log.Warningln(r.Headers)
				log.Warningln(string(r.Body))
			}
			_, e = c.AccountingFields.RealTimeDebitsResponse.ToJSON()
			if e != nil {
				replyError = e.Error()
			}

		}
	})
	collector.OnScraped(func(r *colly.Response) {
		//fmt.Println("Finished", r.Request.URL)
	})
	collector.OnError(func(r *colly.Response, err error) {
		replyError = "Error: " + r.Request.URL.String() + err.Error()
	})
	collector.OnRequest(func(r *colly.Request) {
		for _, val := range c.tableTokens {
			r.Headers.Set(val.name, val.value)
		}
		r.Headers.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		r.Headers.Set("X-Requested-With", "XMLHttpRequest")
		r.Headers.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		r.Headers.Set("Origin", efilingURL)
		r.Headers.Set("Referer", efilingURL+accountingInterestPath)

		fmt.Println("Visiting", r.URL)
	})
	collector.RedirectHandler = redirHandler
	//TODO use timeout
	collector.UserAgent = userAgent
	err = collector.Post(efilingURL+accountingRealTimeDebits, c.AccountingFields.RealTimeDebitsPostData())
	if err != nil {
		return err
	} else if replyError != "" {
		return errors.New(replyError)
	}

	c.collector = collector.Clone()
	return nil
}
