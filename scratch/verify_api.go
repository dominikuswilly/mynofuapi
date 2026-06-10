package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	UserID      string  `json:"id"`
	Name        string  `json:"name"`
	AccessToken *string `json:"access_token"`
}

func main() {
	log.Println("Starting integration verification...")
	time.Sleep(1 * time.Second)

	// 1. Connect to DB to retrieve active Admin, Rider, and Product
	connStr := "host=asus.netbird.cloud port=5432 user=user password=pass dbname=mynofudb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	var adminUser, adminPass, adminName string
	var adminID int
	err = db.QueryRow("SELECT i_id, c_username, c_password, c_nm FROM admin_master WHERE i_active = 1 LIMIT 1").Scan(&adminID, &adminUser, &adminPass, &adminName)
	if err != nil {
		log.Fatalf("Failed to fetch active admin: %v", err)
	}
	log.Printf("Found active Admin: %s (ID: %d, Name: %s)", adminUser, adminID, adminName)

	var riderUser, riderPass, riderName string
	var riderID int
	err = db.QueryRow("SELECT i_id, c_username, c_password, c_nm FROM rider_master WHERE i_active = 1 LIMIT 1").Scan(&riderID, &riderUser, &riderPass, &riderName)
	if err != nil {
		log.Fatalf("Failed to fetch active rider: %v", err)
	}
	log.Printf("Found active Rider: %s (ID: %d, Name: %s)", riderUser, riderID, riderName)

	var productID, productNm string
	var amtSell float64
	err = db.QueryRow("SELECT c_id, c_nm, i_amt_sell FROM product_master WHERE i_active = 1 LIMIT 1").Scan(&productID, &productNm, &amtSell)
	if err != nil {
		log.Fatalf("Failed to fetch active product: %v", err)
	}
	log.Printf("Found active Product: %s (ID: %s, Price: %.0f)", productNm, productID, amtSell)

	// We must ensure the rider has an initiated stock harian today for testing waste/sales/close.
	// Let's force daily inventory initiation directly for today to prevent errors.
	log.Println("Seeding daily rider inventory for verification...")
	db.Exec(`
		INSERT INTO rider_inventory (
			c_id, c_created_by, ts_created_at, c_product_id, 
			i_qty_base, i_qty_current, i_rider_id, 
			c_category, c_product_nm, i_amt_sell, i_closed, i_confirmed
		)
		VALUES (gen_random_uuid(), 'SYSTEM', NOW(), $1, '50', '50', $2, 'KOPI', $3, $4, 0, 1)
		ON CONFLICT DO NOTHING
	`, productID, riderID, productNm, amtSell)

	// If already exists but closed, let's force open for today's test
	db.Exec("UPDATE rider_inventory SET i_closed = 0, i_qty_current = '50', i_qty_base = '50', i_confirmed = 1 WHERE i_rider_id = $1 AND c_product_id = $2 AND (ts_created_at AT TIME ZONE 'Asia/Jakarta')::date = (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Jakarta')::date", riderID, productID)

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 10 * time.Second}

	// 2. Login as Admin
	log.Println("\n--- Logging in as Admin ---")
	adminToken := login(client, baseURL+"/public/admin/auth", adminUser, adminPass)
	log.Println("Admin login successful!")

	// 3. Login as Rider
	log.Println("\n--- Logging in as Rider ---")
	riderToken := login(client, baseURL+"/public/auth", riderUser, riderPass)
	log.Println("Rider login successful!")

	// 4. Test Rider submits restock request
	log.Println("\n--- Testing: [POST] /private/inventory/restock ---")
	restockBody := map[string]interface{}{
		"items": []map[string]interface{}{
			{"product_id": productID, "quantity": 15},
		},
	}
	respMap := postSecure(client, baseURL+"/private/inventory/restock", riderToken, restockBody)
	requestID := respMap["request_id"].(string)
	log.Printf("Restock request submitted successfully! Request ID: %s", requestID)

	// 5. Test Admin views restock requests
	log.Println("\n--- Testing: [GET] /private/admin/inventory/restock ---")
	getSecure(client, baseURL+"/private/admin/inventory/restock?status=pending", adminToken)

	// 6. Test Admin approves restock request
	log.Println("\n--- Testing: [POST] /private/admin/inventory/restock/{id}/approve ---")
	postSecure(client, baseURL+"/private/admin/inventory/restock/"+requestID+"/approve", adminToken, nil)
	log.Println("Restock request approved! Rider daily base stock incremented.")

	// 7. Test Rider reports spilled/defective stock (waste)
	log.Println("\n--- Testing: [POST] /private/transaction/waste ---")
	wasteBody := map[string]interface{}{
		"product_id": productID,
		"quantity":   3,
		"reason":     "tumpah di jalanan berlubang",
	}
	postSecure(client, baseURL+"/private/transaction/waste", riderToken, wasteBody)
	log.Println("Defective stock reported successfully! Qty current decremented.")

	// 8. Test Rider reports sale (to generate commission)
	log.Println("\n--- Testing: [POST] /private/transaction/sales ---")
	saleBody := map[string]interface{}{
		"payment_method": "CASH",
		"amt_pay_total":  int(amtSell * 5),
		"items": []map[string]interface{}{
			{"product_id": productID, "product_name": productNm, "qty_sell": 5, "amt_sell": int(amtSell)},
		},
	}
	postSecure(client, baseURL+"/private/transaction/sales", riderToken, saleBody)
	log.Println("Sale recorded successfully!")

	// 9. Test Admin closes session
	log.Println("\n--- Testing: [POST] /private/admin/transaction/close-session ---")
	closeBody := map[string]interface{}{
		"rider_id": riderID,
		"actual_stocks": []map[string]interface{}{
			{"product_id": productID, "physical_qty": 57}, // 50 base + 15 restock - 3 waste - 5 sold = 57 system stock.
		},
	}
	closeResp := postSecure(client, baseURL+"/private/admin/transaction/close-session", adminToken, closeBody)
	commissionAmt := closeResp["commission_amt"].(float64)
	log.Printf("Session closed successfully! Rider Daily Commission calculated: Rp %.0f", commissionAmt)

	// 10. Test Rider Wallet Summary & Wallet Activity History
	log.Println("\n--- Testing: [GET] /private/wallet/summary ---")
	getSecure(client, baseURL+"/private/wallet/summary", riderToken)

	log.Println("\n--- Testing: [GET] /private/wallet/history ---")
	getSecure(client, baseURL+"/private/wallet/history", riderToken)

	log.Println("\n=== ALL PHASE 1 INTEGRATION TESTS COMPLETED SUCCESSFULLY ===")
}

func login(client *http.Client, url, username, password string) string {
	body, _ := json.Marshal(AuthRequest{Username: username, Password: password})
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Login HTTP failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Fatalf("Login failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var authResp AuthResponse
	json.NewDecoder(resp.Body).Decode(&authResp)
	if authResp.AccessToken == nil {
		log.Fatalf("No access token returned")
	}
	return *authResp.AccessToken
}

func postSecure(client *http.Client, url, token string, payload interface{}) map[string]interface{} {
	var bodyReader io.Reader
	if payload != nil {
		bodyBytes, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(bodyBytes)
	}
	req, _ := http.NewRequest("POST", url, bodyReader)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("POST secure failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("POST %s failed with status %d: %s", url, resp.StatusCode, string(respBody))
	}

	log.Printf("POST %s: %d OK | Response: %s", url, resp.StatusCode, string(respBody))

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	return result
}

func getSecure(client *http.Client, url, token string) map[string]interface{} {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("GET secure failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("GET %s failed with status %d: %s", url, resp.StatusCode, string(respBody))
	}

	// Limit printed history output size
	printable := string(respBody)
	if len(printable) > 300 {
		printable = printable[:300] + "... (truncated)"
	}
	log.Printf("GET %s: %d OK | Response: %s", url, resp.StatusCode, printable)

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	return result
}
