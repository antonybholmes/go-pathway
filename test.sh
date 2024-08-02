curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"geneset":{"name": "test","genes":["ITGAL"]},"datasets":["signaturedb.v2022"]}' \
  http://localhost:8080/modules/pathway/test
