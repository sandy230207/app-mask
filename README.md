# app-mask
## DB schema
![](https://i.imgur.com/NHs0w1x.png)

### USER 使用者
```
USER 使用者
+--------+--------------+------+-----+---------+----------------+
| Field  | Type         | Null | Key | Default | Extra          |
+--------+--------------+------+-----+---------+----------------+
| id     | int          | NO   | PRI | NULL    | auto_increment |
| pid    | varchar(10)  | NO   |     | NULL    |                |
| name   | varchar(10)  | NO   |     | NULL    |                |
| passwd | varchar(256) | NO   |     | NULL    |                |
+--------+--------------+------+-----+---------+----------------+
```
- id: 使用者代號，**使用者不會知道此資訊**
- pid: 身分證字號，用作帳號
- name: 使用者姓名
- passwd: 使用者密碼
> id 具唯一性、pid 具唯一性

### STORE 店家
```
STORE 店家
+-------+-------------+------+-----+---------+----------------+
| Field | Type        | Null | Key | Default | Extra          |
+-------+-------------+------+-----+---------+----------------+
| id    | int         | NO   | PRI | NULL    | auto_increment |
| name  | varchar(10) | NO   |     | NULL    |                |
+-------+-------------+------+-----+---------+----------------+
```
- id: 店家代號
- name: 店家名稱
> id 具唯一性

### INVENTORY 存貨
```
INVENTORY 存貨
+----------+------+------+-----+---------+----------------+
| Field    | Type | Null | Key | Default | Extra          |
+----------+------+------+-----+---------+----------------+
| id       | int  | NO   | PRI | NULL    | auto_increment |
| store_id | int  | NO   | MUL | NULL    |                |
| date     | date | NO   |     | NULL    |                |
| stock    | int  | NO   |     | NULL    |                |
+----------+------+------+-----+---------+----------------+
```
- id: 訂單代號
- store_id: 店家代號
- date: 日期
- stock: 存貨量
> id 具唯一性、(store_id, date) 具唯一性

### ORDER 訂單
```
ORDER 訂單
+--------------+------------+------+-----+---------+----------------+
| Field        | Type       | Null | Key | Default | Extra          |
+--------------+------------+------+-----+---------+----------------+
| id           | int        | NO   | PRI | NULL    | auto_increment |
| user_id      | int        | NO   | MUL | NULL    |                |
| inventory_id | int        | NO   | MUL | NULL    |                |
| pick_up      | tinyint(1) | NO   |     | NULL    |                |
+--------------+------------+------+-----+---------+----------------+
```
- id: 存貨代號
- user_id: 使用者代號
- inventory_id: 存貨代號
- pick_up: 是否已取貨
> id 具唯一性、(user_id, inventory_id) 具唯一性

## APIs
- **request key 大小寫(應該)都通**
- **使用 json format**

Result status code:
- `200`: 成功
- `4XX`: 使用者錯誤
- `5XX`: 伺服器錯誤

Host:
- `34.80.188.97:3000` (http)
- example:
    - GET `http://35.194.136.172:3000/api/queryUser`
    - POST `http://35.194.136.172:3000/api/signUp`
        request: `{"pid": "A123456789", "name": "Bob", "passwd": "asdad123414"}
`
### 註冊
POST `/api/signUp`

request:
```json=
{
    "pid": "A123456789",
    "name": "Bob",
    "passwd": "asdad123414"
}
```
response:
```json
{
    "ResultCode": "200",
    "ResultMessage": "Create Successfully!"
}
```

### 登入
POST `/api/signIn`

request:
```json
{
    "pid": "A123456789",
    "passwd": "asdad123414"
}
```
- pid: 身分證字號，用作帳號
- passwd: 使用者密碼

response:
```json
{
    "ResultCode": "200",
    "ResultMessage": {
        "ID": 11,
        "Pid": "",
        "Name": "Bob",
        "Passwd": ""
    }
}
```
> 只會代 `id` 和 `name` 回來
- ID: 使用者代號，**使用者不會知道此資訊**
- Name: 使用者姓名

### 查詢所有商店的資訊
GET `/api/queryStore`

response:
```json
{
    "ResultCode": "200",
    "ResultMessage": [
        {
            "ID": 1,
            "Name": "A"
        },
        {
            "ID": 2,
            "Name": "B"
        }
    ]
}
```
- ID: 店家代號
- Name: 店家名稱

### 查詢某年某月某日的所有店家口罩存貨量(只列出有存貨的店家)
GET `/api/queryStockByDate/{date}`

date 格式：`YY-MM-DD`

URL:
```
/api/queryStockByDate/2020-12-15
```

response:
```json
{
    "ResultCode": "200",
    "ResultMessage": [
        {
            "ID": 1,
            "Name": "A",
            "Stock": 50
        },
        {
            "ID": 2,
            "Name": "B",
            "Stock": 100
        }
    ]
}
```
- ID: 店家代號
- Name: 店家名稱
- Stock: 店家該日存貨量

### 查詢某店家未來(不含今日)口罩存貨量
GET `/api/queryStockByStore/{id}`

id: 店家代號

URL:
```
/api/queryStockByStore/1
```

response:
```json
{
    "ResultCode": "200",
    "ResultMessage": [
        {
            "Date": "2020-12-17",
            "Stock": 100
        },
        {
            "Date": "2020-12-18",
            "Stock": 100
        },
        {
            "Date": "2020-12-19",
            "Stock": 200
        },
        {
            "Date": "2020-12-20",
            "Stock": 250
        }
    ]
}
```
Date: 日期
Stock: 存貨量

### 查詢某使用者的已成立訂單
POST `/api/queryHistoryOrder`

request:
```json
{
    "id": 11
}
```
id: 使用者代號

response:
```json
{
    "ResultCode": "200",
    "ResultMessage": [
        {
            "OrderID": 7,
            "StoreName": "A",
            "Date": "2020-12-17",
            "IsPickUp": false
        }
    ]
}
```
- OrderID: 訂單編號
- StoreName: 店家名稱
- Date: 取貨日期
- IsPickUp: 是否取貨

### 預購
POST `/api/book`

request:
```json
{
    "userid": 11,
    "storeid": 1,
    "date": "2020-12-17"
}
```
response:
```json
{
    "ResultCode": "200",
    "ResultMessage": "The order was created successfully!"
}
```
#### 錯誤 status code
- 訂購日期不是今日以後：`411`
- 距前次購買日期不足 14 天：`412`
- 有已成立但未取消之訂單：`413`
- 存貨量不足：`414`
- 棄單次數已達 3 次：`415`


### 取消訂單
POST `/api/cancelOrder`

request:
```json
{
    "id": 1
}
```

response:
```json
{
    "ResultCode": "200",
    "ResultMessage": "The order was canceled successfully!"
}
```



## 測試用 APIs (之後會移除)

### 查詢所有使用者
GET `/api/queryUser`

URL:
```
/api/queryUser
```
response:
```json
{
    "ResultCode": "200",
    "ResultMessage": [
        ...
        {
            "ID": 9,
            "Pid": "G123456789",
            "Name": "sandy9",
            "Passwd": "abcdefg"
        },
        {
            "ID": 10,
            "Pid": "D1234567",
            "Name": "alice",
            "Passwd": "123414"
        },
        {
            "ID": 11,
            "Pid": "A123456789",
            "Name": "Bob",
            "Passwd": "asdad123414"
        },
        {
            "ID": 12,
            "Pid": "A111156789",
            "Name": "Catherine",
            "Passwd": "testpasswd"
        }
    ]
}
```

### 查詢所有訂單
GET `api/queryOrder`

URL:
```
api/queryOrder
```
response:
```json
{
    "ResultCode": "200",
    "ResultMessage": [
        ...
        {
            "ID": 5,
            "UserID": 10,
            "InventoryID": 1,
            "PickUp": false
        },
        {
            "ID": 6,
            "UserID": 10,
            "InventoryID": 1,
            "PickUp": false
        },
        {
            "ID": 7,
            "UserID": 11,
            "InventoryID": 5,
            "PickUp": false
        },
        {
            "ID": 8,
            "UserID": 11,
            "InventoryID": 6,
            "PickUp": false
        }
    ]
}
```

### 查詢所有存貨
GET `/api/queryInventory`

URL:
```
/api/queryInventory
```
response:
```json
{
    "ResultCode": "200",
    "ResultMessage": [
        ...
        {
            "ID": 6,
            "StoreID": 1,
            "Date": "2020-12-18",
            "Stock": 99
        },
        {
            "ID": 7,
            "StoreID": 1,
            "Date": "2020-12-19",
            "Stock": 200
        },
        {
            "ID": 8,
            "StoreID": 1,
            "Date": "2020-12-20",
            "Stock": 250
        }
    ]
}
```

### 新增存貨
POST `/api/insertInventory`

request:
```json
{
    "storeID": 1,
    "date": "2020-12-21",
    "stock": 300
}
```
- storeID: 店家代號
- date: 日期
- stock: 進貨量

response:
```json
{
    "ResultCode": "200",
    "ResultMessage": "Insert Inventory Successfully!"
}
```

### 新增店家
POST `/api/insertStore`
request:
```json
{
    "name":"C"
}
```
name: 店家名稱
response:
```json
{
    "ResultCode": "200",
    "ResultMessage": "Insert Store Successfully!"
}
```

### 標記訂單為已取貨
POST `/api/pickUp`

request:
```json
{
    "id": 2
}
```
- id: 訂單編號

response:
```json
{
    "ResultCode": "200",
    "ResultMessage": "Pick Up Successfully!"
}
```

## Useful Commands
https://github.com/kubernetes/minikube

## Deploy 時的一些雷

### LoadBalancer Service has pending IP address
If the result after running LoadBalancer like this:
```
NAME                    TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
service/kubernetes      ClusterIP      10.96.0.1        <none>        443/TCP          21m
service/mask-nodeport   NodePort       10.106.131.184   <none>        3306:30006/TCP   8m55s
service/mask-server     LoadBalancer   10.99.118.5      <pending>     3000:32318/TCP   6s
service/mysql           ClusterIP      10.96.1.1        <none>        3306/TCP         9m47s
```
then run `minikube tunnel`

After running this command, service will look like:
```
NAME                    TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
service/kubernetes      ClusterIP      10.96.0.1        <none>        443/TCP          26m
service/mask-nodeport   NodePort       10.106.131.184   <none>        3306:30006/TCP   14m
service/mask-server     LoadBalancer   10.99.118.5      127.0.0.1     3000:32318/TCP   5m26s
service/mysql           ClusterIP      10.96.1.1        <none>        3306/TCP         15m
```

### Error: `Kubernetes cluster unreachable: Get "https://127.0.0.1:32772/version?timeout=32s": dial tcp 127.0.0.1:32772: connect: connection refused`

When running `helm install . --generate-name` may occur this error, then run `minikube start`

因為沒有指定 cluster


###  no matches for kind "Deployment" in version "extensions/v1beta1"
change Deployment and StatefulSet apiVersion to `apiVersion: apps/v1`

### HPA is not working
https://github.com/kubernetes-sigs/metrics-server/issues/41
https://stackoverflow.com/questions/54106725/docker-kubernetes-mac-autoscaler-unable-to-find-metrics


## Note
[Code](https://github.com/sandy230207/app-mask/tree/feature/api)
[Deployment config files](https://github.com/sandy230207/k8s-mask/tree/feature/app)


部署：
到資料夾底下（ex: mask/db底下）
輸入：`helm install . --generate-name`

範例圖：
![](https://i.imgur.com/JaLcWnB.png)

如果失敗(ex：server壞掉)：
輸入：`helm uninstall chart-160377840`

### Test
Windows:
`set REQUEST_URL="http://34.80.188.97:3000/api/signIn"`
`main.exe`

### 進 DB
`kubectl exec -it pod/chart-1610509478-mask-db-7f5d9768cd-hlj24 -- mysql -u root -p`
