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
  local headers=$3
  local data=$4

  # Define a unique delimiter to separate body and status
  local delimiter="__STATUS_DELIMITER__"

  # Perform the curl request
  response=$(curl -s -w "${delimiter}%{http_code}" -X "$method" $headers -d "$data" "$url")

  # Extract body and status code using the delimiter
  response_body=$(echo "$response" | sed "s/${delimiter}.*//")
  response_status=$(echo "$response" | sed "s/.*${delimiter}//")

  echo "$response_body" "$response_status"
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
AUTH_RESPONSE_AND_STATUS=$(perform_curl "POST" "$API_URL/authenticate" "-H 'Content-Type: application/json'" "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

# Split the response into body and status
AUTH_RESPONSE=$(echo "$AUTH_RESPONSE_AND_STATUS" | awk '{print $1}')
AUTH_STATUS=$(echo "$AUTH_RESPONSE_AND_STATUS" | awk '{print $2}')

echo "Auth Response: $AUTH_RESPONSE"  # Debugging line
echo "Auth Status: $AUTH_STATUS"      # Debugging line

check_status $AUTH_STATUS "Authenticate User"

JWT_TOKEN=$(echo "$AUTH_RESPONSE" | jq -r '.token')

if [ -z "$JWT_TOKEN" ] || [ "$JWT_TOKEN" == "null" ]; then
  echo "Error: Failed to get JWT token"
  exit 1
fi

echo "Authentication Successful. JWT Token Obtained."

# 4. Start Test
echo "Starting Test..."
# **Note:** Adjust the payload according to your API's requirements.
# Here, we're assuming that the `start-test` endpoint expects parameters like logRate, logSize, etc.
START_TEST_PAYLOAD=$(cat <<EOF
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

START_TEST_RESPONSE_AND_STATUS=$(perform_curl "POST" "$API_URL/start-test" "-H 'Content-Type: application/json' -H 'Authorization: Bearer $JWT_TOKEN'" "$START_TEST_PAYLOAD")

# Split the response into body and status
START_TEST_RESPONSE=$(echo "$START_TEST_RESPONSE_AND_STATUS" | awk '{print $1}')
START_TEST_STATUS=$(echo "$START_TEST_RESPONSE_AND_STATUS" | awk '{print $2}')

echo "Start Test Response: $START_TEST_RESPONSE"  # Debugging line
echo "Start Test Status: $START_TEST_STATUS"      # Debugging line

check_status $START_TEST_STATUS "Start Test"

TEST_ID=$(echo "$START_TEST_RESPONSE" | jq -r '.testID')

if [ -z "$TEST_ID" ] || [ "$TEST_ID" == "null" ]; then
  echo "Error: Failed to retrieve Test ID from Start Test response"
  exit 1
fi

echo "Test Started Successfully with Test ID: $TEST_ID"

# 5. Schedule Test
echo "Scheduling Test..."
# **Note:** Adjust the payload according to your API's requirements.
# Assuming the API expects a `scheduleTime` field in ISO8601 format.
SCHEDULE_TIME=$(date -u -d "+10 seconds" +"%Y-%m-%dT%H:%M:%SZ")  # Schedule to start after 10 seconds

SCHEDULE_PAYLOAD=$(cat <<EOF
{
  "testID": "$TEST_ID",
  "scheduleTime": "$SCHEDULE_TIME"
}
EOF
)

SCHEDULE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "$SCHEDULE_PAYLOAD" \
  "$API_URL/schedule-test")
check_status $SCHEDULE_STATUS "Schedule Test"
echo "Test Scheduled Successfully with status $SCHEDULE_STATUS"

# 6. Cancel Test
echo "Cancelling Test..."
CANCEL_PAYLOAD="{\"testID\":\"$TEST_ID\"}"
CANCEL_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "$CANCEL_PAYLOAD" \
  "$API_URL/cancel-test")
check_status $CANCEL_STATUS "Cancel Test"
echo "Test Cancelled Successfully with status $CANCEL_STATUS"

# 7. Restart Test
echo "Restarting Test..."
# **Note:** Adjust the payload according to your API's requirements.
RESTART_PAYLOAD=$(cat <<EOF
{
  "testID": "$TEST_ID",
  "logRate": 1,
  "duration": 60
}
EOF
)

RESTART_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "$RESTART_PAYLOAD" \
  "$API_URL/restart-test")
check_status $RESTART_STATUS "Restart Test"
echo "Test Restarted Successfully with status $RESTART_STATUS"

# 8. Save Results
echo "Saving Results..."
# **Note:** Adjust the payload according to your API's requirements.
SAVE_RESULTS_PAYLOAD=$(cat <<EOF
{
  "testID": "$TEST_ID",
  "completedAt": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
)

SAVE_RESULTS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "$SAVE_RESULTS_PAYLOAD" \
  "$API_URL/save-results")
check_status $SAVE_RESULTS_STATUS "Save Results"
echo "Results Saved Successfully with status $SAVE_RESULTS_STATUS"

# 9. Get All Tests
echo "Getting All Tests..."
GET_ALL_TESTS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  "$API_URL/get-all-tests")
check_status $GET_ALL_TESTS_STATUS "Get All Tests"
echo "Retrieved All Tests Successfully with status $GET_ALL_TESTS_STATUS"

# 10. Get Test By ID
echo "Getting Test By ID..."
GET_TEST_BY_ID_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  "$API_URL/tests/$TEST_ID")
check_status $GET_TEST_BY_ID_STATUS "Get Test By ID"
echo "Retrieved Test By ID Successfully with status $GET_TEST_BY_ID_STATUS"

# 11. Metrics
echo "Getting Metrics..."
METRICS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/metrics")
check_status $METRICS_STATUS "Metrics"
echo "Metrics Retrieved Successfully with status $METRICS_STATUS"

echo "All tests passed successfully!"
