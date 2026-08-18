curl -X POST --location "http://localhost:8000/bots" \
    -H "Accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{
          "id": 1,
          "type": "news",
          "query": "situation in iran",
          "last_ran_at": "2026-01-01T05:00:00Z",
          "headless": false
        }'