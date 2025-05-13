package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"

	"github.com/otiai10/gosseract/v2"
)

func main() {
	waitUntil()
	r := NewBrowserRepo("https://tixcraft.com/activity/detail/24_jaychou")
	r.clickBuyTicket()
	r.clickBuyNowButton()
	r.clickFirstAvailableTicket([]string{"5880", "4880", "3880"})
	r.selectTicketQuantity("2")
	r.clickAgreeCheckbox()
	r.waitForCaptchaAndConfirm()
	r.fillCreditCardInfo("cardNumber", "month", "year", "cvv")
}

type BrowserRepo struct {
	Browser *rod.Browser
	Page    *rod.Page
}

func NewBrowserRepo(url string) *BrowserRepo {
	r := new(BrowserRepo)
	// 連接到已運行的瀏覽器
	u := launcher.NewUserMode().MustLaunch()
	r.Browser = rod.New().ControlURL(u).MustConnect()

	// 檢查是否有已開啟的 tixcraft.com 特定活動頁面
	pages, err := r.Browser.Pages()
	if err != nil {
		panic(fmt.Sprintf("獲取分頁失敗: %v", err))
	}

	for _, p := range pages {
		if p.MustInfo().URL == url {
			r.Page = p
			fmt.Println("檢測到已開啟的演唱會活動頁面")
			break
		}
	}

	// 如果沒有找到已開啟的分頁，則開啟一個新的
	if r.Page == nil {
		r.Page = r.Browser.MustPage(url)
		fmt.Println("開啟新的演唱會活動頁面")
	}

	err = r.Page.WaitLoad()
	if err != nil {
		panic(fmt.Sprintf("頁面加載失敗: %v", err))
	}

	return r
}

// 等待到12點
func waitUntil() {
	now := time.Now()
	noon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())

	if now.After(noon) {
		noon = noon.Add(24 * time.Hour)
	}

	duration := noon.Sub(now)
	fmt.Println("等待到12點...")
	time.Sleep(duration)
	fmt.Println("現在是12點!")
}

// 自動識別圖片驗證碼並點擊確認
func (r *BrowserRepo) waitForCaptchaAndConfirm() {
	captchaInput := r.Page.MustElement("#TicketForm_verifyCode")
	captchaInput.MustWaitVisible()

	fmt.Println("正在識別驗證碼...")

	captchaImage := r.Page.MustElement("#TicketForm_verifyCode-image")
	screenshot, err := captchaImage.Screenshot(proto.PageCaptureScreenshotFormatPng, 1)
	if err != nil {
		log.Fatal("截圖驗證碼失敗:", err)
	}

	client := gosseract.NewClient()
	defer client.Close()

	err = client.SetImageFromBytes(screenshot)
	if err != nil {
		log.Fatal("設置圖片失敗:", err)
	}

	text, err := client.Text()
	if err != nil {
		log.Fatal("識別驗證碼失敗:", err)
	}

	captchaCode := strings.TrimSpace(text)
	if len(captchaCode) != 4 {
		log.Fatal("識別的驗證碼長度不正確:", captchaCode)
	}

	captchaInput.MustInput(captchaCode)
	fmt.Println("已自動輸入驗證碼:", captchaCode)

	confirmButton := r.Page.MustElement("button.btn.btn-primary.btn-green")
	confirmButton.MustClick()

	fmt.Println("已點擊確認張數按鈕")

	time.Sleep(500 * time.Millisecond)
}

// 點擊同意
func (r *BrowserRepo) clickAgreeCheckbox() {
	checkbox := r.Page.MustElement("#TicketForm_agree")
	checkbox.MustWaitVisible()
	checkbox.MustClick()
	fmt.Println("已點擊同意複選框")
}

func (r *BrowserRepo) fillCreditCardInfo(cardNumber, expirationMonth, expirationYear, cvv string) {
	// 等待信用卡表單加載完成
	r.Page.MustElement("#cardNumber").MustWaitVisible()

	// 填入卡號
	r.Page.MustElement("#cardNumber").MustInput(cardNumber)
	fmt.Println("已填入信用卡卡號")

	// 選擇到期月份
	r.Page.MustElement("#ExpirationMonth").MustSelect(expirationMonth)
	fmt.Println("已選擇到期月份")

	// 選擇到期年份
	r.Page.MustElement("#ExpirationYear").MustSelect(expirationYear)
	fmt.Println("已選擇到期年份")

	// 填入CVV
	r.Page.MustElement("#check_num").MustInput(cvv)
	fmt.Println("已填入卡片檢查碼")

	// 短暫等待以確保所有輸入都已完成
	time.Sleep(1 * time.Second)
}

func (r *BrowserRepo) clickAgreeAndProceed() {
	// 等待按鈕可見
	r.Page.MustElement("#submitButton").MustWaitVisible()

	// 使用XPath找到並點擊按鈕
	agreeButtonXPath := "//button[@id='submitButton' and contains(text(), '我同意本節目規則，下一步')]"
	agreeButton := r.Page.MustElementX(agreeButtonXPath)

	// 滾動到按鈕可見
	agreeButton.MustScrollIntoView()

	// 短暫等待以確保按鈕可交互
	time.Sleep(500 * time.Millisecond)

	// 點擊按鈕
	agreeButton.MustClick()

	fmt.Println("成功點擊 '我同意本節目規則，下一步' 按鈕")
}

// 選擇票數
func (r *BrowserRepo) selectTicketQuantity(quantity string) {
	r.Page.MustElement("#ticketPriceList").MustWaitVisible()

	selector := r.Page.MustElement("select.form-select.mobile-select[name^='TicketForm[ticketPrice]']")
	selector.MustSelect(quantity)

	time.Sleep(500 * time.Millisecond)

	selectedValue := r.Page.MustEval(`() => {
		const select = document.querySelector("select.form-select.mobile-select[name^='TicketForm[ticketPrice]']");
		return select ? select.value : null;
	}`).String()

	if selectedValue == quantity {
		fmt.Printf("成功選擇 %s 張票\n", quantity)
	} else {
		fmt.Printf("選擇失敗，當前選中的值是: %s\n", selectedValue)
	}
}

// 點擊第一個可用的限制票價
func (r *BrowserRepo) clickFirstAvailableTicket(prices []string) {
	r.Page.MustElement(".area-list").MustWaitVisible()

	var element *rod.Element
	var ticketInfo string

	for _, price := range prices {
		xpathExpression := fmt.Sprintf("//ul[contains(@class, 'area-list')]//a[contains(., '%s')]", price)
		elements, err := r.Page.ElementsX(xpathExpression)
		if err == nil && len(elements) > 0 {
			element = elements[0]
			ticketInfo = element.MustText()
			break
		}
	}

	if element == nil {
		fmt.Println("沒有找到可用的票")
		return
	}

	element.MustClick()
	fmt.Printf("成功點擊票價: %s\n", ticketInfo)
}

// 點擊立即訂購按鈕
func (r *BrowserRepo) clickBuyNowButton() {
	buyNowXPath := "//tr[contains(.//td[1], '2024/12/07')]//button[contains(@class, 'btn-primary') and contains(text(), '立即訂購')]"
	buyNowElement := r.Page.MustElementX(buyNowXPath)
	dataHref := buyNowElement.MustAttribute("data-href")

	if dataHref != nil {
		fmt.Printf("成功獲取 data-href: %s\n", *dataHref)
		r.Page.MustNavigate(*dataHref)
		fmt.Println("在新頁面中打開了連結")
	} else {
		fmt.Println("未找到 data-href 屬性或屬性為空")
	}
}

// 點擊立即購票
func (r *BrowserRepo) clickBuyTicket() {
	buyTicketXPath := "//div[text()='立即購票']"
	for {
		element := r.Page.MustElementX(buyTicketXPath)
		if element.MustVisible() {
			element.MustScrollIntoView()
			element.MustClick()
			fmt.Println("成功點擊 \"立即購票\" 按鈕")
			break
		}
		r.Page.MustReload()
		time.Sleep(100 * time.Millisecond)
	}
}
