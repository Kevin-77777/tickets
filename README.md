🎫 周杰倫演唱會搶票test (Golang + Rod + OCR)

這是一個使用 Golang 開發的 Tixcraft 自動購票工具，透過 Rod 操作瀏覽器，並結合 OCR 自動識別驗證碼，快速搶購演唱會門票。

🔧 功能簡介
- 自動開啟 Tixcraft 指定活動頁面
- 等待指定時間（例如中午 12 點）自動開啟搶票流程
- 自動點擊「立即購票」與「立即訂購」
- 自動選擇可用票價與張數
- 自動勾選同意條款
- OCR 自動辨識圖片驗證碼（透過 [gosseract](https://github.com/otiai10/gosseract)）
- 支援後續信用卡資料填寫