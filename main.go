package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"

    "github.com/gofiber/fiber/v2"
)

type ApiResponse struct {
    StatusCode int         `json:"statusCode"`
    Status     bool        `json:"status"`
    Message    string      `json:"message"`
    Data       struct {
        Hasil  string `json:"hasil"`
        DataSp DataSp `json:"data_sp"`
    } `json:"data"`
}

type DataSp struct {
    Prefix       Value         `json:"prefix"`
    Status4G     Value         `json:"status_4g"`
    Dukcapil     Value         `json:"dukcapil"`
    ActiveCard   Value         `json:"active_card"`
    ActivePeriod Value         `json:"active_period"`
    GracePeriod  Value         `json:"grace_period"`
    Quotas       QuotasWrapper `json:"quotas"`
}

type Value struct {
    Value string `json:"value"`
}

type QuotasWrapper struct {
    Value []Quota `json:"value"`
}

type Quota struct {
    Name        string        `json:"name"`
    DateEnd     string        `json:"date_end"`
    DetailQuota []DetailQuota `json:"detail_quota"`
}

type DetailQuota struct {
    Name          string `json:"name"`
    DataType      string `json:"data_type"`
    TotalText     string `json:"total_text"`
    RemainingText string `json:"remaining_text"`
}

func main() {
    app := fiber.New()

    app.Get("/cek-kuota", func(c *fiber.Ctx) error {
        nomorHP := c.Query("nomor_hp")
        if nomorHP == "" {
            return c.Status(400).JSON(fiber.Map{
                "error": "Nomor HP diperlukan!",
            })
        }

        apiUrl := fmt.Sprintf("https://apigw.kmsp-store.com/sidompul/v3/cek_kuota?msisdn=%s&isJSON=true", nomorHP)

        req, _ := http.NewRequest("GET", apiUrl, nil)
        req.Header.Set("Authorization", "Basic c2lkb21wdWxhcGk6YXBpZ3drbXNw")
        req.Header.Set("X-API-Key", "4352ff7d-f4e6-48c6-89dd-21c811621b1c")
        req.Header.Set("X-App-Version", "3.0.0")
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{
                "error": err.Error(),
            })
        }
        defer resp.Body.Close()

        bodyBytes, err := io.ReadAll(resp.Body)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{
                "error": "Gagal membaca body",
            })
        }

        fmt.Println("=== RAW RESPONSE ===")
        fmt.Println(string(bodyBytes))
        fmt.Println("====================")

        if resp.StatusCode != http.StatusOK {
            return c.Status(resp.StatusCode).JSON(fiber.Map{
                "error":  "Gagal mengambil data",
                "status": resp.StatusCode,
                "raw":    string(bodyBytes),
            })
        }

        var apiResp ApiResponse
        err = json.Unmarshal(bodyBytes, &apiResp)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{
                "error": "Gagal parsing data",
                "raw":   string(bodyBytes),
            })
        }

        data := apiResp.Data.DataSp

        hasil := fmt.Sprintf(`📃 RESULT:

MSISDN: %s

Tipe Kartu: %s
Status 4G: %s
Status Dukcapil: %s
Umur Kartu: %s
Masa Aktif: %s
Masa Berakhir Tenggang: %s
===========================

`, nomorHP, data.Prefix.Value, data.Status4G.Value, data.Dukcapil.Value, data.ActiveCard.Value, data.ActivePeriod.Value, data.GracePeriod.Value)

        for _, q := range data.Quotas.Value {
            hasil += fmt.Sprintf(`🎁 Quota: %s
🍂 Aktif Hingga: %s
===========================

`, q.Name, q.DateEnd)

            for _, dq := range q.DetailQuota {
                hasil += fmt.Sprintf(`🎁 Benefit: %s
🎁 Tipe Kuota: %s
🎁 Kuota: %s
🌲 Sisa Kuota: %s

`, dq.Name, dq.DataType, dq.TotalText, dq.RemainingText)
            }
        }

        return c.JSON(fiber.Map{
            "statusCode": apiResp.StatusCode,
            "status":     apiResp.Status,
            "message":    apiResp.Message,
            "data": fiber.Map{
                "hasil":   hasil,
                "data_sp": data,
            },
        })
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "3000"
    }
    log.Fatal(app.Listen(":" + port))
}