#!/bin/bash

# variabels

export AUTH_TOKEN=

export URL_HOST=https://go-api-global.architecture.caradhras.io/inventory

#-----------------------------------------------------
URL_GET="${URL_HOST}/info"

STATUS_CODE=$(curl -s -w " HTTP:%{http_code}" "$URL_GET" \
	--header "Content-Type: application/json" \
	--header "Authorization: $AUTH_TOKEN ")

if echo "$STATUS_CODE" | grep -q "HTTP:200"; then
	echo "HTTP:200 GET /info"
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
  	echo "HTTP:200 GET /inventory/product/MOBILE-${RANDOM_INV}"
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
  	echo "HTTP:200 PUT inventory/product/MOBILE-${RANDOM_INV}"
else
	echo -e "\e[31m** ERROR $STATUS_CODE ==> inventory/product/MOBILE-${RANDOM_INV}\e[0m"
fi

#------------------------------------------
RANDOM_INV=$((RANDOM % 999 + 1))

URL_POST="${URL_HOST}/product"
PAYLOAD='{"sku":"WINE-'$RANDOM_INV'", "type":"BEVERAGE", "name": "WINE '$RANDOM_INV'", "status":"IN-STOCK"}'

STATUS_CODE=$(curl -s -w " HTTP:%{http_code}" "$URL_POST" \
	--header "Content-Type: application/json" \
	--header "Authorization: $AUTH_TOKEN" \
	--data "$PAYLOAD")

if echo "$STATUS_CODE" | grep -q "HTTP:200"; then
  	echo "HTTP:200 POST inventory/product/WINE-${RANDOM_INV}"
else
	echo -e "\e[31m** ERROR $STATUS_CODE ==> /product\e[0m"
fi

#------------------------------------------
RANDOM_INV=$((RANDOM % 999 + 1))

URL_POST="${URL_HOST}/product"
PAYLOAD='{"sku":"BEER-'$RANDOM_INV'", "type":"BEVERAGE", "name": "BEER '$RANDOM_INV'", "status":"IN-STOCK"}'

STATUS_CODE=$(curl -s -w " HTTP:%{http_code}" "$URL_POST" \
	--header "Content-Type: application/json" \
	--header "Authorization: $AUTH_TOKEN" \
	--data "$PAYLOAD")

if echo "$STATUS_CODE" | grep -q "HTTP:200"; then
  	echo "HTTP:200 POST inventory/product/BEER-${RANDOM_INV}"
else
	echo -e "\e[31m** ERROR $STATUS_CODE ==> /product\e[0m"
fi