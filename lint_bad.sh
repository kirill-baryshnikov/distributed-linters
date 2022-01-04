curl curl -v -X POST -H "Content-Type: application/json" \
    -d '{"content": "int main(){a =b}"}' \
    http://127.0.0.1:10000/v1/lint/python
