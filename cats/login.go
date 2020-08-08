package cats

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

type loginPostFields struct {
	LoginURL                 string
	Username                 string
	Password                 string
	AcceptTermsAndConditions string
	LanguageID               string
}

type CatsUserData struct {
	LogedInTimeStamp    string
	Cookie              string
	Username            string
	Password            string
	loginRequiredFields *loginPostFields
	AccountingFields    *accountingFields
	collector           *colly.Collector
	formToken           *tokenFromHTML
	ajaxToken           *tokenFromHTML
	tableTokens         []*tokenFromHTML
	antiForgeryToken    tokenFromHTML
}

func (c *CatsUserData) loginHasAllFields() bool {
	return (c.loginRequiredFields.LoginURL != "") && (c.loginRequiredFields.Username != "") &&
		(c.loginRequiredFields.Password != "") &&
		(c.formToken.name != "") &&
		(c.formToken.value != "") &&
		(c.loginRequiredFields.AcceptTermsAndConditions != "") &&
		(c.loginRequiredFields.LanguageID != "")
}
func (l *loginPostFields) String() string {
	return fmt.Sprintf("%+v", *l)
}
func (c *CatsUserData) loginPostData(username string, password string) map[string]string {
	return map[string]string{
		c.loginRequiredFields.Username:                 username,
		c.loginRequiredFields.Password:                 password,
		c.formToken.name:                               c.formToken.value,
		c.loginRequiredFields.AcceptTermsAndConditions: "true",
		c.loginRequiredFields.LanguageID:               "1",
	}
}

func (c *CatsUserData) String() string {
	var val string
	for _, v := range c.tableTokens {
		val += fmt.Sprintf("%+v", v)
	}

	val += fmt.Sprintf("%+v\n%s", c.ajaxToken, val)
	return fmt.Sprintf("%+v\n%+v\n%s", *c, c.loginRequiredFields, val)
}

func (c *CatsUserData) accountingRealTimeDebitsHasAllFields() bool {
	return len(c.tableTokens) != 2
}
func (c *CatsUserData) accountingDebtsPaymentHasAllFields() bool {
	return (len(c.tableTokens) != 2) && (c.antiForgeryToken.name != "") && (c.antiForgeryToken.value != "")
}

func NewCatsUserData() *CatsUserData {
	return &CatsUserData{
		collector:           colly.NewCollector(),
		loginRequiredFields: &loginPostFields{},
		tableTokens:         make([]*tokenFromHTML, 0),
		ajaxToken:           &tokenFromHTML{},
		formToken:           &tokenFromHTML{},
		antiForgeryToken:    tokenFromHTML{},
		AccountingFields: &accountingFields{
			RealTimeDebitsResponse: &realTimeDebitsResponse{},
			DebtsPayment:           &debtsResponse{},
		},
	}
}

func redirHandler(req *http.Request, via []*http.Request) error {
	fmt.Println("redirect", req.URL)
	return nil
}

func (c *CatsUserData) checkLoginForm() error {
	var loginFormError string = ""
	c.collector.OnHTML(".trim_class", func(e *colly.HTMLElement) {
		if (e.Attr("id") == "UserName") && (e.Attr("name") == "UserName") {
			c.loginRequiredFields.Username = "UserName"
		}
	})
	c.collector.OnHTML("form", func(e *colly.HTMLElement) {
		if (e.Attr("method") == "post") && (e.ChildAttr("input", "name") == "__RequestVerificationToken") {
			c.formToken.name = e.ChildAttr("input", "name")
			c.formToken.value = e.ChildAttr("input", "value")
			c.loginRequiredFields.LoginURL = e.Request.AbsoluteURL(e.Attr("action"))
		}
	})
	c.collector.OnHTML(".trim_class", func(e *colly.HTMLElement) {
		if (e.Attr("id") == "UserName") && (e.Attr("name") == "UserName") {
			c.loginRequiredFields.Username = "UserName"
		}
	})
	c.collector.OnHTML("select", func(e *colly.HTMLElement) {
		if (e.Attr("id") == "LanguageID") && (e.Attr("name") == "LanguageID") {
			c.loginRequiredFields.LanguageID = "LanguageID"
		}
	})
	c.collector.OnHTML("input", func(e *colly.HTMLElement) {
		if (e.Attr("id") == "Password") && (e.Attr("name") == "Password") {
			c.loginRequiredFields.Password = e.Attr("name")
		}
		if (e.Attr("id") == "AcceptTermsAndConditions") && (e.Attr("name") == "AcceptTermsAndConditions") {
			c.loginRequiredFields.AcceptTermsAndConditions = e.Attr("name")
		}
	})
	c.collector.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			loginFormError = r.Request.URL.String() + " status code != 200"
			log.Warningln(loginFormError)
			log.Warningln(r.Request.URL)
			log.Warningln(r.Request.Marshal())
			log.Warningln(r.StatusCode)
			log.Warningln(r.Headers)
			log.Warningln(string(r.Body))
		}
	})
	c.collector.OnScraped(func(r *colly.Response) {
		//fmt.Println("Finished", r.Request.URL)
	})
	c.collector.OnError(func(r *colly.Response, err error) {
		loginFormError = "checkLoginForm " + err.Error()
	})
	c.collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	c.collector.RedirectHandler = redirHandler

	//TODO use timeout
	c.collector.UserAgent = userAgent
	err := c.collector.Visit(loginPath)
	if err != nil {
		return err
	} else if loginFormError != "" {
		return fmt.Errorf("login form, %s", loginFormError)
	} else if !c.loginHasAllFields() {
		return fmt.Errorf("login form, could nURLot find all fields %s", c.loginRequiredFields)
	}
	return nil
}

func (c *CatsUserData) Login() error {
	err := c.checkLoginForm()
	if err != nil {
		return err
	}
	var loginError string
	collector := c.collector.Clone()
	collector.OnHTML(".validation-summary-errors", func(e *colly.HTMLElement) {
		loginError = e.ChildText("span")
	})
	collector.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			loginError = "error login status code != 200"
			log.Warningln(loginError)
			log.Warningln(r.Request.URL)
			log.Warningln(r.Request.Marshal())
			log.Warningln(r.StatusCode)
			log.Warningln(r.Headers)
			log.Warningln(string(r.Body))
		} else {
			token, er := getAjaxToken(r.Body)
			if er != nil {
				loginError = "login does not have ajax request token" + er.Error()
				log.Warningln(loginError)
				log.Warningln(r.Request.URL)
				log.Warningln(r.Request.Marshal())
				log.Warningln(r.StatusCode)
				log.Warningln(r.Headers)
				log.Warningln(string(r.Body))
			} else {
				c.ajaxToken = token
			}
		}
	})
	collector.OnScraped(func(r *colly.Response) {
		//fmt.Println("Finished", r.Request.URL)
	})
	collector.OnError(func(r *colly.Response, err error) {
		loginError = "Error Login: " + r.Request.URL.String() + " " + err.Error()
	})
	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	collector.RedirectHandler = redirHandler

	//TODO use timeout
	collector.UserAgent = userAgent
	err = collector.Post(c.loginRequiredFields.LoginURL, c.loginPostData(c.Username, c.Password))
	if err != nil {
		return err
	} else if loginError != "" {
		return errors.New(loginError)
	}
	c.LogedInTimeStamp = strconv.FormatInt(time.Now().Unix(), 10)
	c.collector = collector.Clone()
	return nil
}

func (c *CatsUserData) IsAuthenticated() (bool, error) {
	isAuth := false
	var loginError string
	collector := c.collector.Clone()

	collector.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			loginError = "error login status code != 200"
			log.Warningln(loginError)
			log.Warningln(r.Request.URL)
			log.Warningln(r.Request.Marshal())
			log.Warningln(r.StatusCode)
			log.Warningln(r.Headers)
			log.Warningln(string(r.Body))
		} else {
			body := string(r.Body)
			if body == `{"IsAuthenticated":true}` {
				isAuth = true
			} else {
				isAuth = false
			}
		}
	})
	collector.OnScraped(func(r *colly.Response) {
		//fmt.Println("Finished", r.Request.URL)
	})
	collector.OnError(func(r *colly.Response, err error) {
		loginError = "Error Login: " + r.Request.URL.String() + " " + err.Error()
		log.Warningln(loginError)
		log.Warningln(r.Request.URL)
		log.Warningln(r.Request.Marshal())
		log.Warningln(r.StatusCode)
		log.Warningln(r.Headers)
		log.Warningln(string(r.Body))
	})
	collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set(c.ajaxToken.name, c.ajaxToken.value)
		fmt.Println("Visiting", r.URL)
	})
	collector.RedirectHandler = redirHandler

	//TODO use timeout
	collector.UserAgent = userAgent
	err := collector.Visit(efilingURL + isAuthPath + c.LogedInTimeStamp)
	if err != nil {
		return false, err
	} else if loginError != "" {
		return false, errors.New(loginError)
	}
	c.collector = collector.Clone()
	return isAuth, nil
}
