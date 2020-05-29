### BUILD/Deploy
```shell script
# build
env GOOS=linux go build main.go
scp main ubuntu@$BANKING_REMOTE:

# deploy
ssh ubuntu@$BANKING_REMOTE
# on remote
export DB_HOST=<dbHost>
./main >trace.log 2>error.log
```
---
### API

#### GET /accounts/{id}
gets account balance  
returning following json:
```json
{
    "id": 55,
    "balance": 101.58
}
```

Returns `404` if account with `id` does not exist

#### POST /accounts/
Create new account, send with body:
```json
{
    "id": 55,
    "balance": 101.58
}
```
returns created account info:
```json
{
    "id": 55,
    "balance": 101.58
}
```

Returns `400` if account with `id` exists or balance is negative

#### POST /ops/transfer
Transfer amount from one account to another
```json
{
    "fromId": 55,
    "toId":66,
    "amount": 0.1
}
```

returns new balances/unchanged balances in case of error

```json
{
    "from": {
        "id": 55,
        "balance": 102.22
    },
    "to": {
        "id": 66,
        "balance": 97.84
    }
}
```

Returns `400` if `fromId == toId` or amount is negative

---
### Load testing

Using `loadtest` tool  
install using `npm install -g loadtest`
```shell script
export BANKING_URL=http://$BANKING_REMOTE:3333

cd ./loadTesting
# Test transferring between two fixed accounts randomly continuously
loadtest -n 10000 --rps 101 -c 5 -R deadLockCatcher.js $BANKING_URL

# Test all API endpoints with random IDs (should return only 2xx)
loadtest -n 10000 --rps 101 -c 5 -R generalNoErrTest.js $BANKING_URL

# Test all endpoints with random IDs 
# (should return some 4xx errors, amount could be negative, acc_ids could exist or be equal)
loadtest -n 10000 --rps 101 -c 5 -R generalTest.js $BANKING_URL
```