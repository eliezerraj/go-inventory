#!/bin/bash

# variabels

export AUTH_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b2tlbl91c2UiOiJhY2Nlc3MiLCJpc3MiOiJsYW1iZGEtZ28taWRlbnRpZHkiLCJ2ZXJzaW9uIjoiMS4yIiwidXNlcm5hbWUiOiJhZG1pbi0xMDEiLCJqd3RfaWQiOiI1NjgzMDA1Yy1hOTBlLTQzMGYtOTE5MS1lYzdhYmQ3ZGQxOGUiLCJraWQiOiJhdXRoLWtleTpzZXJ2ZXItcHVibGljLmtleSIsInRpZXIiOiJ0aWVyMSIsImFwaV9hY2Nlc3Nfa2V5IjoiQVBJX0FDQ0VTU19LRVlfQURNSU5fMDAxIiwic2NvcGUiOlsidG9vbDppbmZvIiwidG9vbDpoZWFsdGgiLCJ0b29sOmdldF9wcm9kdWN0IiwiYWRtaW4iXSwiZXhwIjoxNzc0NzExMjc4fQ.XBpNTn5PC9eQcmDt9traLwp2ZlDO_QbZN8ZVpnLcp-U
export URL_HOST=https://go-api-global.architecture.caradhras.io/inventory
#export URL_HOST=http://localhost:7000

PRODUCT="cheese-fr"
TYPE="beverage"

RANDOM_INV=$((RANDOM % 30 + 1))

SKU="${PRODUCT}-${RANDOM_INV}"
NAME="${PRODUCT} ${RANDOM_INV}"

echo "------------------------------"
echo  "sku": "${SKU} type: ${TYPE} name: ${NAME}"
echo "------------------------------"

#----------------------- INFO------------------------------
URL_GET="${URL_HOST}/info"

STATUS_CODE=$(curl -s -w " HTTP:%{http_code}" "$URL_GET" \
	--header "Content-Type: application/json" \
	--header "Authorization: $AUTH_TOKEN ")

if echo "$STATUS_CODE" | grep -q "HTTP:200"; then
	echo "HTTP:200 GET /info"
else
	echo -e "\e[31m** ERROR $STATUS_CODE ==> /info\e[0m"
fi

# --------------------- POST inventory/product------------------------------
URL_POST="${URL_HOST}/product"

PAYLOAD=$(cat <<EOF
	{
		"sku": "${SKU}",
		"type": "${TYPE}",
		"name": "${NAME}",
		"status": "IN-STOCK",
		"lead_time": 20
	}
EOF
)

STATUS_CODE=$(curl -s -w " HTTP:%{http_code}" "$URL_POST" \
	--header "Content-Type: application/json" \
	--header "Authorization: $AUTH_TOKEN" \
	--data "$PAYLOAD")

if echo "$STATUS_CODE" | grep -q "HTTP:200"; then
  	echo "HTTP:200 POST inventory/product/${SKU}"
elif echo "$STATUS_CODE" | grep -q "HTTP:400"; then
  	echo -e "\e[38;2;255;165;0m** ERROR $STATUS_CODE ==> /product\e[0m"
else
	echo -e "\e[31m** ERROR $STATUS_CODE ==> /product\e[0m"
fi

# --------------------- GET inventory/product------------------------------
URL_GET="${URL_HOST}/inventory/product/${SKU}"

STATUS_CODE=$(curl -s -w " HTTP:%{http_code}" "$URL_GET" \
	--header "Content-Type: application/json" \
	--header "Authorization: $AUTH_TOKEN ")

if echo "$STATUS_CODE" | grep -q "HTTP:200"; then
  	echo "HTTP:200 GET /inventory/product/${SKU}"
elif echo "$STATUS_CODE" | grep -q "HTTP:404"; then	
	echo -e "\e[38;2;255;165;0m** ERROR $STATUS_CODE ==> /inventory/product/${SKU}\e[0m"
else
	echo -e "\e[31m** ERROR $STATUS_CODE ==> /inventory/product/${SKU}\e[0m"
fi

# --------------------- PUT inventory/product------------------------------
URL_PUT="${URL_HOST}/inventory/product/${SKU}"
PAYLOAD='{"available":0, "reserved":0, "sold": 0}'

STATUS_CODE=$(curl -s -w " HTTP:%{http_code}" -X PUT "$URL_PUT" \
	--header "Content-Type: application/json" \
	--header "Authorization: $AUTH_TOKEN" \
	--data "$PAYLOAD")

if echo "$STATUS_CODE" | grep -q "HTTP:200"; then
  	echo "HTTP:200 PUT inventory/product/${SKU}"
elif echo "$STATUS_CODE" | grep -q "HTTP:404"; then	
	echo -e "\e[38;2;255;165;0m** ERROR $STATUS_CODE ==> inventory/product/${SKU}\e[0m"	
else
	echo -e "\e[31m** ERROR $STATUS_CODE ==> inventory/product/${SKU}\e[0m"
fi

