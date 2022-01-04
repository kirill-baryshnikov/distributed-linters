curl -v -X POST -H "Content-Type: application/json" \
    -d '{"version": "bin/python-linter-1.0"}' \
    http://127.0.0.1:10000/v1/admin/rollback/python
