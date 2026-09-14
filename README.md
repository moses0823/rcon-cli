# rcontui

一個基於終端機的 RCON 用戶端，提供互動式 TUI 介面，以及可選的 Secure RCON 驗證功能。

本專案 Fork 自 [gorcon/rcon-cli](https://github.com/gorcon/rcon-cli)，原專案採用 MIT License 發布。

本 Fork 在保留原專案授權與版權聲明的基礎上，加入額外功能與修改。

## 功能

* 互動式終端機 TUI
* Minecraft RCON 控制台
* 多伺服器設定
* Secure RCON 驗證
* HMAC-SHA256 Challenge-Response 驗證
* 獨立的驗證 Secret 檔案
* Windows 支援

## 安裝

下載 `rcontui.exe`，並將其放置在獨立的目錄中：

```text
C:\rcontui\
├── rcontui.exe
├── rcon.yaml
└── secrets\
    └── 6b7t-moses.key
```

## 設定

建立 `rcon.yaml`：

```yaml
servers:
  6b7t:
    address: "abula.tw:25576"
    security:
      enabled: true
      client-id: "moses"
      secret-file: "secrets/6b7t-moses.key"
```

### 設定選項

| 選項                     | 說明                       |
| ---------------------- | ------------------------ |
| `address`              | RCON 或 Secure RCON 伺服器位址 |
| `security.enabled`     | 啟用 Secure RCON           |
| `security.client-id`   | 用戶端識別名稱                  |
| `security.secret-file` | 驗證 Secret 檔案路徑           |

## Secure RCON

啟用後，`rcontui` 會連接 Secure RCON Gateway，而不是直接連接 Minecraft 原生 RCON。

```text
rcontui
   │
   │ Secure authentication
   ▼
Secure RCON Gateway
   │
   ▼
Minecraft Native RCON
```

範例：

```yaml
address: "abula.tw:25576"

security:
  enabled: true
  client-id: "moses"
  secret-file: "secrets/6b7t-moses.key"
```

Minecraft 原生 RCON Port 應保持私有，不應直接暴露在 Internet 上。

## Secret 檔案

Secret 會獨立儲存：

```text
secrets\
└── 6b7t-moses.key
```

請勿將 Secret 檔案提交至 Git。

請勿公開分享 Secret 檔案。

## 執行

開啟 PowerShell：

```powershell
cd C:\rcontui
.\rcontui.exe
```

使用鍵盤進行操作：

```text
↑ / ↓     選擇伺服器
Enter     連線
Q         離開
```

## 多伺服器

可以設定多個伺服器：

```yaml
servers:
  6b7t:
    address: "abula.tw:25576"
    security:
      enabled: true
      client-id: "moses"
      secret-file: "secrets/6b7t-moses.key"

  local:
    address: "127.0.0.1:25575"
    security:
      enabled: false
```

每個伺服器都可以獨立啟用或停用 Secure RCON。

## 安全性

Secure RCON 使用 HMAC-SHA256 Challenge-Response 驗證。

驗證過程使用：

* Client ID
* Server 產生的 Nonce
* 基於時間的驗證
* 共用 Secret
* HMAC-SHA256

共用 Secret 不會在驗證過程中直接傳送。

## 建置

本專案需要 Go。

執行測試：

```powershell
go test ./...
```

建立 Windows 執行檔：

```powershell
go build -o rcontui.exe .
```

## Fork 資訊

本 Repository 基於：

**原始專案：** `gorcon/rcon-cli`

**原始 Repository：** https://github.com/gorcon/rcon-cli

原始專案採用 MIT License。

本 Fork 保留原專案的版權聲明與授權。

## License

本專案採用 MIT License 發布。

完整授權內容請參閱 [`LICENSE`](LICENSE)。

原始專案及其相關版權聲明仍受其原始授權條款約束。
