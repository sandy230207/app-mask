package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"net/http"

	"app-mask/services"
)

var DbAddress = "test:test@tcp(10.98.0.40:3306)/MASK"

// const DbAddress = "test:test@/MASK"

type ApiResponse struct {
	ResultCode    string
	ResultMessage interface{}
}

type User struct {
	ID     int32
	Pid    string
	Name   string
	Passwd string
}

type OrderForm struct {
	ID          int32
	UserID      int32
	InventoryID int32
	PickUp      bool
}

type StoreStock struct {
	ID    int
	Name  string
	Stock int
}

type Store struct {
	ID   int
	Name string
}

type StockOfDate struct {
	Date  string
	Stock int
}

type HistoryOrder struct {
	OrderID   int
	StoreName string
	Date      string
	IsPickUp  bool
}

type Order struct {
	UserID  int
	StoreID int
	Date    string
}

type Inventory struct {
	ID      int32
	StoreID int32
	Date    string
	Stock   int32
}

func InitDBAddress(ip string) {
	DbAddress = "test:test@tcp(" + ip + ":3306)/MASK"
}

// 登入
func SignIn(w http.ResponseWriter, r *http.Request) {
	var user User
	var userData User
	var resMsg ApiResponse

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024)) //io.LimitReader限制大小
	if err != nil {
		resMsg.ResultCode = "413"
		resMsg.ResultMessage = err // Response body size should be less than 1KB.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	err = json.Unmarshal(body, &user)
	if err != nil {
		resMsg.ResultCode = "400"
		resMsg.ResultMessage = err // Failed to parse into json
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer r.Body.Close()

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Server cannot connect to database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	var (
		id   sql.NullInt32
		name sql.NullString
	)
	err = db.QueryRow("SELECT id,name FROM USER WHERE pid=? AND passwd=?", user.Pid, user.Passwd).Scan(&id, &name)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to query data from database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	if !id.Valid || !name.Valid {
		resMsg.ResultCode = "403"
		resMsg.ResultMessage = err // User not found. 查無帳號or密碼(帳號or密碼輸入錯誤)
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		return
	}
	db.Close()

	userData.ID = id.Int32
	userData.Name = name.String
	resMsg = ApiResponse{"200", userData}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}

// 註冊
func SignUp(w http.ResponseWriter, r *http.Request) {
	var user User
	var userIDList []int
	var resMsg ApiResponse

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024)) //io.LimitReader限制大小
	if err != nil {
		resMsg.ResultCode = "413"
		resMsg.ResultMessage = err // Response body size should be less than 1KB.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	err = json.Unmarshal(body, &user)
	if err != nil {
		resMsg.ResultCode = "400"
		resMsg.ResultMessage = err // Failed to parse into json
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer r.Body.Close()

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Server cannot connect to database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	rows, err := db.Query("SELECT id FROM USER WHERE pid=?", user.Pid)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to query data from database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			resMsg.ResultCode = "500"
			resMsg.ResultMessage = err // Failed to scan data.
			services.ResponseWithJson(w, http.StatusOK, resMsg)
			log.Println(err)
			return
		}
		userIDList = append(userIDList, id)
		log.Printf("id: %v\n", id)
	}

	if userIDList != nil {
		resMsg.ResultCode = "400"
		resMsg.ResultMessage = "Error: User existed."
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(resMsg.ResultMessage)
		return
	}

	stmt, err := db.Prepare("INSERT USER SET pid=?,name=?,passwd=?")
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to insert data to database(1).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	_, err = stmt.Exec(user.Pid, user.Name, user.Passwd)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to insert data to database(2).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	db.Close()

	resMsg = ApiResponse{"200", "Create Successfully!"}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}

// 查詢某日所有店家口罩存量(只列出有存量的店家)
func QueryStockByDate(w http.ResponseWriter, r *http.Request) {
	var date string

	var storeStockList []StoreStock
	var resMsg ApiResponse

	vars := mux.Vars(r)
	if _, ok := vars["date"]; !ok {
		resMsg.ResultCode = "400"
		resMsg.ResultMessage = "URL should be followed by the 'date' you want to query."
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println("URL should be followed by the 'date'.")
		return
	}
	date = vars["date"]

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Server cannot connect to database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	rows, err := db.Query("SELECT STORE.id, name, stock FROM INVENTORY JOIN STORE ON store_id=STORE.id WHERE date=? AND stock>0", date)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to query data from database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id    int
			name  string
			stock int
		)
		if err := rows.Scan(&id, &name, &stock); err != nil {
			resMsg.ResultCode = "500"
			resMsg.ResultMessage = err // Failed to scan data.
			services.ResponseWithJson(w, http.StatusOK, resMsg)
			log.Println(err)
			return
		}
		storeStockList = append(storeStockList, StoreStock{id, name, stock})
		log.Printf("id: %v, name: %v, stock: %v\n", name, id, stock)
	}
	db.Close()

	resMsg = ApiResponse{"200", storeStockList}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}

// 查詢所有店家的資訊
func QueryStore(w http.ResponseWriter, r *http.Request) {

	var storeList []Store
	var resMsg ApiResponse

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Server cannot connect to database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	rows, err := db.Query("SELECT * FROM STORE")
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to query data from database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id   int
			name string
		)
		if err := rows.Scan(&id, &name); err != nil {
			resMsg.ResultCode = "500"
			resMsg.ResultMessage = err // Failed to scan data.
			services.ResponseWithJson(w, http.StatusOK, resMsg)
			log.Println(err)
			return
		}
		storeList = append(storeList, Store{id, name})
		log.Printf("id: %v, name: %v\n", id, name)
	}
	db.Close()

	resMsg = ApiResponse{"200", storeList}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}

// 查詢某店家未來一星期口罩存量
func QueryStockByStore(w http.ResponseWriter, r *http.Request) {
	var storeID int
	var stockList []StockOfDate
	var resMsg ApiResponse

	vars := mux.Vars(r)
	if _, ok := vars["id"]; !ok {
		resMsg.ResultCode = "400"
		resMsg.ResultMessage = "URL should be followed by the 'id' which is the storeID you want to query."
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println("URL should be followed by the 'id'.")
		return
	}
	storeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		resMsg.ResultCode = "400"
		resMsg.ResultMessage = err // 'id' should be int.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Server cannot connect to database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	rows, err := db.Query("SELECT date, stock FROM INVENTORY WHERE store_id=? AND date>CURDATE()", storeID)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to query data from database."
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			date  string
			stock int
		)
		if err := rows.Scan(&date, &stock); err != nil {
			resMsg.ResultCode = "500"
			resMsg.ResultMessage = err // Failed to scan data.
			services.ResponseWithJson(w, http.StatusOK, resMsg)
			log.Println(err)
			return
		}
		stockList = append(stockList, StockOfDate{date, stock})
		log.Printf("date: %v, stock: %v\n", date, stock)
	}
	db.Close()

	resMsg = ApiResponse{"200", stockList}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}

// 查詢歷史訂單
func QueryHistoryOrder(w http.ResponseWriter, r *http.Request) {
	var user User
	var historyOrder []HistoryOrder
	var resMsg ApiResponse

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024)) //io.LimitReader限制大小
	if err != nil {
		resMsg.ResultCode = "413"
		resMsg.ResultMessage = err // Response body size should be less than 1KB.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	err = json.Unmarshal(body, &user)
	if err != nil {
		resMsg.ResultCode = "400"
		resMsg.ResultMessage = err // Failed to parse into json
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer r.Body.Close()

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Server cannot connect to database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	rows, err := db.Query("SELECT ORDER.id, name, date, pick_up FROM MASK.ORDER JOIN INVENTORY ON inventory_id=INVENTORY.id JOIN STORE on store_id=STORE.id WHERE user_id=?", user.ID)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to query data from database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			orderID   int
			storeName string
			date      string
			isPickUp  bool
		)
		if err := rows.Scan(&orderID, &storeName, &date, &isPickUp); err != nil {
			resMsg.ResultCode = "500"
			resMsg.ResultMessage = err // Failed to scan data.
			services.ResponseWithJson(w, http.StatusOK, resMsg)
			log.Println(err)
			return
		}
		historyOrder = append(historyOrder, HistoryOrder{orderID, storeName, date, isPickUp})
		log.Printf("orderID: %v, storeName: %v, date: %v, isPickUp: %v\n", orderID, storeName, date, isPickUp)
	}
	db.Close()

	resMsg = ApiResponse{"200", historyOrder}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}

// 查詢前次購買日期
func queryLastBuyDate(userID int) (string, error) {
	var inventoryIDList []int
	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		return "", err
	}
	var date sql.NullString
	// 12/01 buy
	//
	rows, err := db.Query("SELECT inventory_id FROM MASK.ORDER WHERE user_id=? AND pick_up=true", userID)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var inventoryID int
		if err := rows.Scan(&inventoryID); err != nil {
			return "", nil
		}
		inventoryIDList = append(inventoryIDList, inventoryID)
		log.Printf("inventoryID: %v\n", inventoryID)
	}

	var maxInventoryID = -1
	for _, v := range inventoryIDList {
		if v > maxInventoryID {
			maxInventoryID = v
		}
	}
	err = db.QueryRow("SELECT date FROM MASK.INVENTORY WHERE id=?", maxInventoryID).Scan(&date)
	if err != nil {
		return "2020-01-01", err // 查無存貨編號=-1(從未購買過)
	}
	if !date.Valid {
		log.Println("date is null.")
		return "", nil // 查無存貨編號=-1(從未購買過)
	}
	log.Printf("last buy date: %v\n", date.String)
	db.Close()

	return date.String, nil
}

// 查詢某店家某日存量編號及口罩存量
func queryStockByStoreAndDate(storeID int, date string) (int32, int32, error) {
	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		return -1, -1, err
	}
	var (
		id    sql.NullInt32
		stock sql.NullInt32
	)
	err = db.QueryRow("SELECT id, stock FROM INVENTORY WHERE store_id=? AND date=?", storeID, date).Scan(&id, &stock)
	if err != nil {
		return -1, -1, err
	}
	if !id.Valid || !stock.Valid {
		log.Println("id is null or stock is null.")
		return -1, -1, nil // 該日無存量
	}
	db.Close()

	return id.Int32, stock.Int32, nil
}

// 計算沒有領貨的次數
// > 3 return false
// <= 3 return true
func countPickUp(userID int) (int32, error) {
	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		return -1, err
	}
	var count sql.NullInt32
	err = db.QueryRow("SELECT count(id) FROM MASK.ORDER WHERE user_id=? AND pick_up=false", userID).Scan(&count)
	if err != nil {
		return -1, err
	}
	if !count.Valid {
		log.Println("count is null")
		return -1, nil //?????
	}
	db.Close()

	return count.Int32, nil
}

func computeDate(date string) (float64, error) {

	currentTime := time.Now().Format("2006-01-02")
	log.Println("currentTime: ", currentTime)

	year, err := strconv.Atoi(strings.Split(currentTime, "-")[0])
	if err != nil {
		return 0, err
	}
	month, err := strconv.Atoi(strings.Split(currentTime, "-")[1])
	if err != nil {
		return 0, err
	}
	day, err := strconv.Atoi(strings.Split(currentTime, "-")[2])
	if err != nil {
		return 0, err
	}
	t1 := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	log.Println("lastDate: ", date)

	year, err = strconv.Atoi(strings.Split(date, "-")[0])
	if err != nil {
		return 0, err
	}
	month, err = strconv.Atoi(strings.Split(date, "-")[1])
	if err != nil {
		return 0, err
	}
	day, err = strconv.Atoi(strings.Split(date, "-")[2])
	if err != nil {
		return 0, err
	}
	t2 := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	log.Printf("current time: %v, param time %v\n", t1, t2)
	log.Println("current - param: ", t1.Sub(t2))

	timeDifference := t1.Sub(t2).Hours()
	return timeDifference, nil
}

// 查詢已成立且取貨日期為今日以後之訂單是否存在
func queryOrderByUserID(userID int) (bool, error) {
	var inventoryID sql.NullInt32
	var date sql.NullString

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		return true, err
	}

	err = db.QueryRow("SELECT MAX(inventory_id) FROM MASK.ORDER WHERE user_id=?", userID).Scan(&inventoryID)
	if err != nil {
		return true, err
	}
	if !inventoryID.Valid {
		return false, nil // 未訂貨過
	}

	err = db.QueryRow("SELECT date FROM INVENTORY WHERE id=?", inventoryID.Int32).Scan(&date)
	if err != nil {
		return true, err
	}
	if !date.Valid {
		return false, errors.New("The inventory has no date data.")
	}

	db.Close()

	timeDifference, err := computeDate(date.String)
	if err != nil {
		return true, err
	}
	if timeDifference <= 0 {
		return true, errors.New("You have a book in the next _ day or now in future.")
	}

	return false, nil
}

// 登記預購
func Book(w http.ResponseWriter, r *http.Request) {
	var order Order
	var resMsg ApiResponse

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024)) //io.LimitReader限制大小
	if err != nil {
		resMsg.ResultCode = "413"
		resMsg.ResultMessage = err // Response body size should be less than 1KB.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	err = json.Unmarshal(body, &order)
	if err != nil {
		resMsg.ResultCode = "400"
		resMsg.ResultMessage = err // Failed to parse into json
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer r.Body.Close()

	// 訂購日期是否大於現在時間
	timeDifference, err := computeDate(order.Date)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	if timeDifference >= 0 {
		resMsg.ResultCode = "411"
		resMsg.ResultMessage = "Cannot book the mask when the date <= now."
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Printf("Now - date = %v\n", timeDifference)
		return
	}

	// 前次購買日期是否 >= 14 天
	lastBuyDate, err := queryLastBuyDate(order.UserID)
	if err != nil && lastBuyDate == "" {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	log.Printf("lastBuyDate: %v\n", lastBuyDate)

	timeDifference, err = computeDate(lastBuyDate)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	if timeDifference < 336 {
		resMsg.ResultCode = "412"
		resMsg.ResultMessage = "Now - lastBuyDate < 336 hr (14 days)"
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println("Now - lastBuyDate < 336 hr (14 days)")
		return
	}

	// 是否有已成立但未取消之訂單
	if ok, err := queryOrderByUserID(order.UserID); err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	} else if ok {
		resMsg.ResultCode = "413"
		resMsg.ResultMessage = "You have an existed order."
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println("You have an existed order.")
		return
	}

	// 存量是否足夠
	inventoryID, stock, err := queryStockByStoreAndDate(order.StoreID, order.Date)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	if inventoryID == -1 || stock == -1 {
		resMsg.ResultCode = "414"
		resMsg.ResultMessage = "The store has no inventory on that day."
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println("The store has no inventory on that day.")
		return
	}
	log.Printf("inventoryID: %v, stock: %v\n", inventoryID, stock)

	// 是否棄單超過 3 次
	count, err := countPickUp(order.UserID)
	if err != nil {
		resMsg.ResultCode = "502"
		resMsg.ResultMessage = err
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	if count > 3 {
		resMsg.ResultCode = "415"
		resMsg.ResultMessage = "User has not picked up the mask more than 3 times."
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println("The store has no inventory on that day.")
		return
	}
	log.Printf("count: %v\n", count)

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Server cannot connect to database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	// 新增訂單
	stmt, err := db.Prepare("INSERT MASK.ORDER SET user_id=?,inventory_id=?,pick_up=?")
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to insert data to database(1).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	_, err = stmt.Exec(order.UserID, inventoryID, false)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to insert data to database(2).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	// 減少該日存貨
	stmt, err = db.Prepare("UPDATE INVENTORY SET stock=? WHERE id=?")
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to update data to database(1).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	_, err = stmt.Exec(stock-1, inventoryID)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to update data to database(2).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	db.Close()

	resMsg = ApiResponse{"200", "The order was created successfully!"}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}

// 取消預購
func CancelOrder(w http.ResponseWriter, r *http.Request) {
	var order OrderForm
	var inventory Inventory
	var resMsg ApiResponse

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024)) //io.LimitReader限制大小
	if err != nil {
		resMsg.ResultCode = "413"
		resMsg.ResultMessage = err // Response body size should be less than 1KB.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	err = json.Unmarshal(body, &order)
	if err != nil {
		resMsg.ResultCode = "400"
		resMsg.ResultMessage = err // Failed to parse into json
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer r.Body.Close()

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Server cannot connect to database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	// 查詢該訂單資料
	order, err = queryOrderByID(order.ID)

	// 查詢該存貨日期&存貨量
	inventory, err = queryInventoryByID(order.InventoryID)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	log.Printf("inventoryID: %v, date: %v, stock: %v\n", order.InventoryID, inventory.Date, inventory.Stock)

	timeDifference, err := computeDate(inventory.Date)
	fmt.Println(inventory.Date)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	if timeDifference >= 0 {
		resMsg.ResultCode = "400"
		resMsg.ResultMessage = "Cannot cancel the order which the date > now."
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Printf("Now - timeDifference = %v\n", timeDifference)
		return
	}

	// 刪除訂單
	stmt, err := db.Prepare("DELETE FROM MASK.ORDER WHERE id=?")
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to insert data to database(1).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	_, err = stmt.Exec(order.ID)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to insert data to database(2).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	// 增加該日存貨
	stmt, err = db.Prepare("UPDATE INVENTORY SET stock=? WHERE id=?")
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to update data to database(1).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	_, err = stmt.Exec(inventory.Stock+1, order.InventoryID)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to update data to database(2).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	db.Close()

	resMsg = ApiResponse{"200", "The order was canceled successfully!"}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}

// 查指定的 Order 的資訊
func queryOrderByID(orderID int32) (OrderForm, error) {
	var order OrderForm
	log.Println("orderID: ", orderID)

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		return order, err
	}

	var (
		id          sql.NullInt32
		userID      sql.NullInt32
		inventoryID sql.NullInt32
		pickUp      sql.NullBool
	)

	err = db.QueryRow("SELECT * FROM MASK.ORDER WHERE id=?", orderID).Scan(&id, &userID, &inventoryID, &pickUp)
	if err != nil {
		return order, err
	}
	if !id.Valid || !userID.Valid || !inventoryID.Valid || !pickUp.Valid {
		return order, nil
	}
	db.Close()
	log.Println("order param: ", id, userID, inventoryID, pickUp)
	log.Println("order: ", order)
	order = OrderForm{id.Int32, userID.Int32, inventoryID.Int32, pickUp.Bool}
	return order, err
}

// 查指定的 Inventory 的資訊
func queryInventoryByID(inventoryID int32) (Inventory, error) {
	var inventory Inventory

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		return inventory, err
	}

	var (
		id      sql.NullInt32
		storeID sql.NullInt32
		date    sql.NullString
		stock   sql.NullInt32
	)

	err = db.QueryRow("SELECT * FROM INVENTORY WHERE id=?", inventoryID).Scan(&id, &storeID, &date, &stock)
	if err != nil {
		return inventory, err
	}
	if !id.Valid || !storeID.Valid || !date.Valid || !stock.Valid {
		return inventory, nil
	}
	db.Close()

	inventory = Inventory{id.Int32, storeID.Int32, date.String, stock.Int32}

	return inventory, err
}

// 以下為測試用 API
// 查詢所有user的資訊
func QueryUser(w http.ResponseWriter, r *http.Request) {
	var userList []User
	var resMsg ApiResponse

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Server cannot connect to database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	rows, err := db.Query("SELECT * FROM USER")
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to query data from database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id     int32
			pid    string
			name   string
			passwd string
		)
		if err := rows.Scan(&id, &pid, &name, &passwd); err != nil {
			resMsg.ResultCode = "500"
			resMsg.ResultMessage = err // Failed to scan data.
			services.ResponseWithJson(w, http.StatusOK, resMsg)
			log.Println(err)
			return
		}
		userList = append(userList, User{id, pid, name, passwd})
		log.Printf("id: %v, pid: %v, name: %v, passwd: %v\n", id, pid, name, passwd)
	}
	db.Close()

	resMsg = ApiResponse{"200", userList}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}

// 查詢所有order的資訊
func QueryOrder(w http.ResponseWriter, r *http.Request) {

	var orderList []OrderForm
	var resMsg ApiResponse

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Server cannot connect to database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	rows, err := db.Query("SELECT * FROM MASK.ORDER")
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to query data from database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id          int32
			userID      int32
			inventoryID int32
			pickUp      bool
		)
		if err := rows.Scan(&id, &userID, &inventoryID, &pickUp); err != nil {
			resMsg.ResultCode = "500"
			resMsg.ResultMessage = err // Failed to scan data.
			services.ResponseWithJson(w, http.StatusOK, resMsg)
			log.Println(err)
			return
		}
		orderList = append(orderList, OrderForm{id, userID, inventoryID, pickUp})
		log.Printf("id: %v, userID: %v, inventoryID: %v, pickUp: %v\n", id, userID, inventoryID, pickUp)
	}
	db.Close()

	resMsg = ApiResponse{"200", orderList}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}

// 查詢所有inventory的資訊
func QueryInventory(w http.ResponseWriter, r *http.Request) {
	var inventoryList []Inventory
	var resMsg ApiResponse

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Server cannot connect to database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	rows, err := db.Query("SELECT * FROM INVENTORY")
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to query data from database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id      int32
			storeID int32
			date    string
			stock   int32
		)
		if err := rows.Scan(&id, &storeID, &date, &stock); err != nil {
			resMsg.ResultCode = "500"
			resMsg.ResultMessage = err // Failed to scan data.
			services.ResponseWithJson(w, http.StatusOK, resMsg)
			log.Println(err)
			return
		}
		inventoryList = append(inventoryList, Inventory{id, storeID, date, stock})
		log.Printf("id: %v, storeID: %v, date: %v, stock: %v\n", id, storeID, date, stock)
	}
	db.Close()

	resMsg = ApiResponse{"200", inventoryList}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}

// 店家 API
// 增加存貨
func InsertInventory(w http.ResponseWriter, r *http.Request) {
	var inventory Inventory
	var inventoryIDList []int
	var resMsg ApiResponse

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024)) //io.LimitReader限制大小
	if err != nil {
		resMsg.ResultCode = "413"
		resMsg.ResultMessage = err // Response body size should be less than 1KB.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	err = json.Unmarshal(body, &inventory)
	if err != nil {
		resMsg.ResultCode = "400"
		resMsg.ResultMessage = err // Failed to parse into json
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer r.Body.Close()

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Server cannot connect to database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}

	rows, err := db.Query("SELECT id FROM INVENTORY WHERE store_id=? AND date=?", inventory.StoreID, inventory.Date)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to query data from database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			resMsg.ResultCode = "500"
			resMsg.ResultMessage = err // Failed to scan data.
			services.ResponseWithJson(w, http.StatusOK, resMsg)
			log.Println(err)
			return
		}
		inventoryIDList = append(inventoryIDList, id)
		log.Printf("id: %v\n", id)
	}

	if inventoryIDList != nil {
		resMsg.ResultCode = "400"
		resMsg.ResultMessage = "Error: Inventory had been inserted already."
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(resMsg.ResultMessage)
		return
	}

	stmt, err := db.Prepare("INSERT INVENTORY SET store_id=?,date=?,stock=?")
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to insert data to database(1).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	_, err = stmt.Exec(inventory.StoreID, inventory.Date, inventory.Stock)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to insert data to database(2).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	db.Close()

	resMsg = ApiResponse{"200", "Insert Inventory Successfully!"}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}

// 增加店家
func InsertStore(w http.ResponseWriter, r *http.Request) {
	var store Store
	var resMsg ApiResponse

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024)) //io.LimitReader限制大小
	if err != nil {
		resMsg.ResultCode = "413"
		resMsg.ResultMessage = err // Response body size should be less than 1KB.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	err = json.Unmarshal(body, &store)
	if err != nil {
		resMsg.ResultCode = "400"
		resMsg.ResultMessage = err // Failed to parse into json
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer r.Body.Close()

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Server cannot connect to database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	stmt, err := db.Prepare("INSERT STORE SET name=?")
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to insert data to database(1).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	_, err = stmt.Exec(store.Name)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to insert data to database(2).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	db.Close()

	resMsg = ApiResponse{"200", "Insert Store Successfully!"}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}

// 取貨
func PickUp(w http.ResponseWriter, r *http.Request) {
	var order OrderForm
	var resMsg ApiResponse

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024)) //io.LimitReader限制大小
	if err != nil {
		resMsg.ResultCode = "413"
		resMsg.ResultMessage = err // Response body size should be less than 1KB.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	err = json.Unmarshal(body, &order)
	if err != nil {
		resMsg.ResultCode = "400"
		resMsg.ResultMessage = err // Failed to parse into json
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	defer r.Body.Close()

	db, err := sql.Open("mysql", DbAddress)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Server cannot connect to database.
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	stmt, err := db.Prepare("UPDATE MASK.ORDER SET pick_up=? WHERE id=?")
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to update data to database(1).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	_, err = stmt.Exec(true, order.ID)
	if err != nil {
		resMsg.ResultCode = "500"
		resMsg.ResultMessage = err // Failed to update data to database(2).
		services.ResponseWithJson(w, http.StatusOK, resMsg)
		log.Println(err)
		return
	}
	db.Close()

	resMsg = ApiResponse{"200", "Pick Up Successfully!"}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}

// Health check
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	var resMsg ApiResponse

	resMsg = ApiResponse{"200", "OK!"}
	services.ResponseWithJson(w, http.StatusOK, resMsg)
}
