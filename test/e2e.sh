#!/bin/bash

echo "=== Starting E2E tests ==="

API="http://localhost:8080"

function section() {
  echo ""
  echo "----------------------------------"
  echo "$1"
  echo "----------------------------------"
}

# 1. team/add
section "1) Create team backend"
curl -s -X POST $API/team/add -H "Content-Type: application/json" -d '{
  "team_name": "backend",
  "members": [
    {"user_id":"u1","username":"Alice","is_active":true},
    {"user_id":"u2","username":"Bob","is_active":true},
    {"user_id":"u3","username":"Eve","is_active":true}
  ]
}'

# 2. team/get
section "2) Get team backend"
curl -s "$API/team/get?team_name=backend"

# 3. pullRequest/create
section "3) Create PR pr1"
curl -s -X POST $API/pullRequest/create -H "Content-Type: application/json" -d '{
  "author_id": "u1",
  "pull_request_id": "pr1",
  "pull_request_name": "fix bug"
}'

# 4. users/getReview
section "4) getReview for u2"
curl -s "$API/users/getReview?user_id=u2"

# 5. pullRequest/merge
section "5) Merge pr1"
curl -s -X POST $API/pullRequest/merge -H "Content-Type: application/json" -d '{
  "pull_request_id":"pr1"
}'

# 6. pullRequest/reassign (should fail - merged)
section "6) Reassign on merged PR (expected error)"
curl -s -X POST $API/pullRequest/reassign -H "Content-Type: application/json" -d '{
  "id":"pr1",
  "reviewerId":"u2"
}'

# 7. Create second PR
section "7) Create PR pr2"
curl -s -X POST $API/pullRequest/create -H "Content-Type: application/json" -d '{
  "author_id": "u1",
  "pull_request_id": "pr2",
  "pull_request_name": "feature A"
}'

# 8. Reassign reviewer on open PR
section "8) Reassign reviewer on pr2"
curl -s -X POST $API/pullRequest/reassign -H "Content-Type: application/json" -d '{
  "id":"pr2",
  "reviewerId":"u2"
}'

# 9. Deactivate team
section "9) Deactivate team backend"
curl -s -X POST $API/team/deactivate -H "Content-Type: application/json" -d '{
  "team": "backend"
}'

# 10. Stats
section "10) Stats"
curl -s "$API/stats"

echo ""
echo "=== E2E DONE ==="