#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Set variables
API_URL="http://localhost:8080"
USERNAME="testuser_$(date +%s)"
EMAIL="${USERNAME}@example.com"
PASSWORD="testpassword"
TEST_ID=""
JWT_TOKEN=""

# Function to check the HTTP response status
check_status() {
  local status_code=$1
  local step=$2
  if [ "$status_code" -ne 200 ] && [ "$status_code" -ne 201 ]; then
    echo "Error: $step failed with status $status_code"
    exit 1
  fi
}

# Function to perform curl and capture response body and status code
perform_curl() {
  local method=$1
  local url=$2
  shift 2
  local headers=("$@")

  # Perform the curl request with a timeout of 30 seconds
  response=$(curl -s -o response_body.txt -w "%{http_code}" -X "$method" "${headers[@]}" "$url" --max-time 30)

  # Capture the status code
  status_code=$(echo "$response" | tail -n1)

  # Read response body from file
  response_body=$(<response_body.txt)

  # Output response body and status code on separate lines
  echo "$response_body"
  echo "$status_code"
}

# Function to calculate future time (supports macOS and Linux)
calculate_future_time() {
  local seconds=$1
  if date --version >/dev/null 2>&1; then
    # GNU date (Linux)
    date -u -d "+${seconds} seconds" +"%Y-%m-%dT%H:%M:%SZ"
  else
    # BSD date (macOS)
    date -u -v +${seconds}S +"%Y-%m-%dT%H:%M:%SZ"
  fi
}

# 1. Health Check
echo "Running Health Check..."
HEALTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/health")
check_status $HEALTH_STATUS "Health Check"
echo "Health Check Passed with status $HEALTH_STATUS"

# 2. Register User
echo "Registering User..."
REGISTER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
  "$API_URL/register")
check_status $REGISTER_STATUS "Register User"
echo "User Registered Successfully with status $REGISTER_STATUS"

# 3. Authenticate User and Get JWT Token
echo "Authenticating User..."
AUTH_OUTPUT=$(curl -s -o auth_response.txt -w "%{http_code}" -X POST "$API_URL/authenticate" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

AUTH_STATUS=$(echo "$AUTH_OUTPUT" | tail -n1)
AUTH_RESPONSE=$(<auth_response.txt)

echo "Auth Response: $AUTH_RESPONSE"
echo "Auth Status: $AUTH_STATUS"

check_status $AUTH_STATUS "Authenticate User"

JWT_TOKEN=$(echo "$AUTH_RESPONSE" | jq -r '.token')

if [ -z "$JWT_TOKEN" ] || [ "$JWT_TOKEN" == "null" ]; then
  echo "Error: Failed to get JWT token"
  exit 1
fi

echo "Authentication Successful. JWT Token Obtained."

# 4. Create Test
echo "Creating Test..."
CREATE_TEST_PAYLOAD=$(cat <<EOF
{
  "logRate": 1,
  "logSize": 100,
  "metricsRate": 1,
  "traceRate": 1,
  "duration": 60,
  "destination": {
    "name": "TestDestination",
    "endpoint": "http://example.com",
    "port": 80,
    "apiKey": "testapikey123"
  }
}
EOF
)

CREATE_TEST_OUTPUT=$(curl -s -o create_test_response.txt -w "%{http_code}" -X POST "$API_URL/create-test" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "$CREATE_TEST_PAYLOAD")

CREATE_TEST_STATUS=$(echo "$CREATE_TEST_OUTPUT" | tail -n1)
CREATE_TEST_RESPONSE=$(<create_test_response.txt)

echo "Create Test Response: $CREATE_TEST_RESPONSE"
echo "Create Test Status: $CREATE_TEST_STATUS"

check_status $CREATE_TEST_STATUS "Create Test"

TEST_ID=$(echo "$CREATE_TEST_RESPONSE" | jq -r '.testID')

if [ -z "$TEST_ID" ] || [ "$TEST_ID" == "null" ]; then
  echo "Error: Failed to retrieve Test ID from Create Test response"
  exit 1
fi

echo "Test Created Successfully with Test ID: $TEST_ID"

# 5. Schedule Test
echo "Scheduling Test..."
SCHEDULE_TIME=$(calculate_future_time 10)

SCHEDULE_PAYLOAD=$(cat <<EOF
{
  "testID": "$TEST_ID",
  "schedule": "$SCHEDULE_TIME"
}
EOF
)

SCHEDULE_OUTPUT=$(curl -s -o schedule_response.txt -w "%{http_code}" -X POST "$API_URL/schedule-test" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "$SCHEDULE_PAYLOAD")

SCHEDULE_STATUS=$(echo "$SCHEDULE_OUTPUT" | tail -n1)
SCHEDULE_RESPONSE=$(<schedule_response.txt)

echo "Schedule Test Response: $SCHEDULE_RESPONSE"
echo "Schedule Test Status: $SCHEDULE_STATUS"

check_status $SCHEDULE_STATUS "Schedule Test"

# 6. Cancel Test
echo "Cancelling Test..."
CANCEL_PAYLOAD=$(cat <<EOF
{
  "testID": "$TEST_ID"
}
EOF
)

CANCEL_OUTPUT=$(curl -s -o cancel_response.txt -w "%{http_code}" -X POST "$API_URL/cancel-test" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "$CANCEL_PAYLOAD")

CANCEL_STATUS=$(echo "$CANCEL_OUTPUT" | tail -n1)
CANCEL_RESPONSE=$(<cancel_response.txt)

echo "Cancel Test Response: $CANCEL_RESPONSE"
echo "Cancel Test Status: $CANCEL_STATUS"

check_status $CANCEL_STATUS "Cancel Test"

# 7. Restart Test
echo "Restarting Test..."
RESTART_PAYLOAD=$(cat <<EOF
{
  "testID": "$TEST_ID",
  "logRate": 1,
  "duration": 60
}
EOF
)

RESTART_OUTPUT=$(curl -s -o restart_response.txt -w "%{http_code}" -X POST "$API_URL/restart-test" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "$RESTART_PAYLOAD")

RESTART_STATUS=$(echo "$RESTART_OUTPUT" | tail -n1)
RESTART_RESPONSE=$(<restart_response.txt)

echo "Restart Test Response: $RESTART_RESPONSE"
echo "Restart Test Status: $RESTART_STATUS"

check_status $RESTART_STATUS "Restart Test"

# 8. Save Results
echo "Saving Results..."
SAVE_RESULTS_PAYLOAD=$(cat <<EOF
{
  "testID": "$TEST_ID",
  "completedAt": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}
EOF
)

SAVE_RESULTS_OUTPUT=$(curl -s -o save_results_response.txt -w "%{http_code}" -X POST "$API_URL/save-results" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "$SAVE_RESULTS_PAYLOAD")

SAVE_RESULTS_STATUS=$(echo "$SAVE_RESULTS_OUTPUT" | tail -n1)
SAVE_RESULTS_RESPONSE=$(<save_results_response.txt)

echo "Save Results Response: $SAVE_RESULTS_RESPONSE"
echo "Save Results Status: $SAVE_RESULTS_STATUS"

check_status $SAVE_RESULTS_STATUS "Save Results"

# 9. Get All Tests
echo "Getting All Tests..."
GET_ALL_TESTS_OUTPUT=$(curl -s -o get_all_tests_response.txt -w "%{http_code}" -X GET "$API_URL/get-all-tests" \
  -H "Authorization: Bearer $JWT_TOKEN")

GET_ALL_TESTS_STATUS=$(echo "$GET_ALL_TESTS_OUTPUT" | tail -n1)
GET_ALL_TESTS_RESPONSE=$(<get_all_tests_response.txt)

echo "Get All Tests Response: $GET_ALL_TESTS_RESPONSE"
echo "Get All Tests Status: $GET_ALL_TESTS_STATUS"

check_status $GET_ALL_TESTS_STATUS "Get All Tests"

# 10. Get Test By ID
echo "Getting Test By ID..."
GET_TEST_BY_ID_OUTPUT=$(curl -s -o get_test_by_id_response.txt -w "%{http_code}" -X GET "$API_URL/tests/$TEST_ID" \
  -H "Authorization: Bearer $JWT_TOKEN")

GET_TEST_BY_ID_STATUS=$(echo "$GET_TEST_BY_ID_OUTPUT" | tail -n1)
GET_TEST_BY_ID_RESPONSE=$(<get_test_by_id_response.txt)

echo "Get Test By ID Response: $GET_TEST_BY_ID_RESPONSE"
echo "Get Test By ID Status: $GET_TEST_BY_ID_STATUS"

check_status $GET_TEST_BY_ID_STATUS "Get Test By ID"

# 11. Metrics
echo "Getting Metrics..."
METRICS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/metrics")
check_status $METRICS_STATUS "Metrics"

echo "Metrics Retrieved Successfully with status $METRICS_STATUS"
echo "All tests passed successfully!"
