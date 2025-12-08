#!/bin/bash

# variabels

export AUTH_TOKEN=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b2tlbl91c2UiOiJhY2Nlc3MiLCJpc3MiOiJsYW1iZGEtZ28taWRlbnRpZHkiLCJ2ZXJzaW9uIjoiMS4wIiwiand0X2lkIjoiNjI1NjZiNjMtNWVjZS00NThmLThjZDUtN2FmNmU2YjE1MThkIiwidXNlcm5hbWUiOiJhZG1pbi0wMDEiLCJ0aWVyIjoidGllcjIiLCJhcGlfYWNjZXNzX2tleSI6IkFQSV9BQ0NFU1NfS0VZX0FETUlOXzAwMSIsInNjb3BlIjpbInRlc3QucmVhZCIsInRlc3Qud3JpdGUiLCJhZG1pbiJdLCJleHAiOjE3NjU2NDkxMzN9.m9so9_dfRfTvErIpOjAZAA1rvdixwypgJJRwpMKj-_5ApL2JA1XphH25tBlJ1HFPmCJ4vBUhQm9U8cj09f3_ZjByDpYCYFUYvrZ9r5YzgZLSV4Os1WUz8RhXrO43zzbuKg9jKqpEtIvtFSIORSCKSUmMvd8_bSdNLmliLy0ZxzY8w4t186W2HmMrqYxG0SdL7hpOYDfFyuHCavNYqG80owrwPCnIU0mZvWbvPO4vAnuuHfeEYnFOiRnblN-LUvHhnF8SfAoeeuCa6yxRKZ1bArkmDaJIYaMWO7ZWLBY7rA-AhGozwYNlqO3_zt4rJFtx0wjXIx5OC4Fq3LoFn2epHda6DFVZeFUamc3zj6W2d05NUVM71gHlqX1UMtMU99lIDhv-rajrddFz0ImQy9_EdgMv-d2FoGVGvZrxi04CycmTWVrPAzKqaJ-kv0fZgv_gjTffFT8FqhuVka7sesWKtWVQjk8gR6K5WGA4SuoLgusxPeTLO_U-5GOgkJSm4tgzbN4zDM3xnzoJorrD87RU9k7H0M6_n2xqJaOugZH7NQcSG9cFCbHdX1jan59Yl3nIJlQzca5Comf_m0sHlvor3Sewt5t8RWe0LKI0UDVTJuMdWPPZRex5sLD19UrVXQ2bjZpuujrQCJ_WYRvqLdoc0dIF8jizmaRQq58KMWpvMWU

export URL_HOST=https://go-api-global.architecture.caradhras.io/inventory

#-----------------------------------------------------
URL_GET="${URL_HOST}/info"

STATUS_CODE=$(curl -s -w " HTTP:%{http_code}" "$URL_GET" \
	--header "Content-Type: application/json" \
	--header "Authorization: $AUTH_TOKEN ")

if echo "$STATUS_CODE" | grep -q "HTTP:200"; then
	echo "HTTP:200 /info"
else
	echo -e "\e[31m** ERROR $STATUS_CODE ==> /info\e[0m"
fi

# ---------------------------------------------------
RANDOM_INV=$((RANDOM % 99 + 1))
URL_GET="${URL_HOST}/inventory/product/MOBILE-${RANDOM_INV}"

STATUS_CODE=$(curl -s -w " HTTP:%{http_code}" "$URL_GET" \
	--header "Content-Type: application/json" \
	--header "Authorization: $AUTH_TOKEN ")

if echo "$STATUS_CODE" | grep -q "HTTP:200"; then
  	echo "HTTP:200 /inventory/product/MOBILE-${RANDOM_INV}"
else
	echo -e "\e[31m** ERROR $STATUS_CODE ==> /inventory/product/MOBILE-${RANDOM_INV}\e[0m"
fi

#------------------------------------------
RANDOM_INV=$((RANDOM % 99 + 1))

URL_PUT="${URL_HOST}/inventory/product/MOBILE-${RANDOM_INV}"
PAYLOAD='{"available":0, "reserved":1, "sold": 1}'

STATUS_CODE=$(curl -s -w " HTTP:%{http_code}" -X PUT "$URL_PUT" \
	--header "Content-Type: application/json" \
	--header "Authorization: $AUTH_TOKEN" \
	--data "$PAYLOAD")

if echo "$STATUS_CODE" | grep -q "HTTP:200"; then
  	echo "HTTP:200 inventory/product/MOBILE-${RANDOM_INV}"
else
	echo -e "\e[31m** ERROR $STATUS_CODE ==> inventory/product/MOBILE-${RANDOM_INV}\e[0m"
fi

#------------------------------------------
#RANDOM_INV=$((RANDOM % 99 + 1))

#URL_POST="${URL_HOST}/product"
#PAYLOAD='{"sku":"MOBILE-'$RANDOM_INV'", "type":"ELETRONIC", "name": "MOBILE '$RANDOM_INV'", "status":"IN-STOCK"}'

#STATUS_CODE=$(curl -s -w " HTTP:%{http_code}" "$URL_POST" \
#	--header "Content-Type: application/json" \
#	--header "Authorization: $AUTH_TOKEN" \
#	--data "$PAYLOAD")

#if echo "$STATUS_CODE" | grep -q "HTTP:200"; then
#  	echo "HTTP:200 inventory/product/MOBILE-${RANDOM_INV}"
#else
#	echo -e "\e[31m** ERROR $STATUS_CODE ==> /product\e[0m"
#fi
