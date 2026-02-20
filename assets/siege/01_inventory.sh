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
elif echo "$STATUS_CODE" | grep -q "HTTP:404"; then	
	echo -e "\e[38;2;255;165;0m** ERROR $STATUS_CODE ==> /inventory/product/MOBILE-${RANDOM_INV}\e[0m"
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
elif echo "$STATUS_CODE" | grep -q "HTTP:404"; then	
	echo -e "\e[38;2;255;165;0m** ERROR $STATUS_CODE ==> inventory/product/MOBILE-${RANDOM_INV}\e[0m"	
else
	echo -e "\e[31m** ERROR $STATUS_CODE ==> inventory/product/MOBILE-${RANDOM_INV}\e[0m"
fi

#------------------------------------------
RANDOM_INV=$((RANDOM % 9999 + 1))

URL_POST="${URL_HOST}/product"
PAYLOAD='{"sku":"GPU-'$RANDOM_INV'", "type":"ELETRONIC", "name": "GPU '$RANDOM_INV'", "status":"IN-STOCK"}'

STATUS_CODE=$(curl -s -w " HTTP:%{http_code}" "$URL_POST" \
	--header "Content-Type: application/json" \
	--header "Authorization: $AUTH_TOKEN" \
	--data "$PAYLOAD")

if echo "$STATUS_CODE" | grep -q "HTTP:200"; then
  	echo "HTTP:200 POST inventory/product/GPU-${RANDOM_INV}"
elif echo "$STATUS_CODE" | grep -q "HTTP:400"; then	
	echo -e "\e[38;2;255;165;0m** ERROR $STATUS_CODE ==> /product\e[0m"	
else
	echo -e "\e[31m** ERROR $STATUS_CODE ==> /product\e[0m"
fi

#------------------------------------------
RANDOM_INV=$((RANDOM % 9999 + 1))

URL_POST="${URL_HOST}/product"
PAYLOAD='{"sku":"LAPTOP-'$RANDOM_INV'", "type":"ELETRONIC", "name": "LAPTOP '$RANDOM_INV'", "status":"IN-STOCK"}'

STATUS_CODE=$(curl -s -w " HTTP:%{http_code}" "$URL_POST" \
	--header "Content-Type: application/json" \
	--header "Authorization: $AUTH_TOKEN" \
	--data "$PAYLOAD")

if echo "$STATUS_CODE" | grep -q "HTTP:200"; then
  	echo "HTTP:200 POST inventory/product/LAPTOP-${RANDOM_INV}"
elif echo "$STATUS_CODE" | grep -q "HTTP:400"; then
  	echo -e "\e[38;2;255;165;0m** ERROR $STATUS_CODE ==> /product\e[0m"
else
	echo -e "\e[31m** ERROR $STATUS_CODE ==> /product\e[0m"
fi