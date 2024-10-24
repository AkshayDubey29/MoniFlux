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

  # Define a unique delimiter to separate body and status
  local delimiter="__STATUS_DELIMITER__"

  # Perform the curl request with a timeout of 30 seconds and capture stderr
  response=$(curl -s -w "${delimiter}%{http_code}" -X "$method" "${headers[@]}" "$url" --max-time 30 2>&1)

  # Check if curl encountered an error (e.g., timeout, DNS failure)
  if [[ "$response" == *"$delimiter"* ]]; then
    # Extract body and status code using the delimiter
    response_body=$(echo "$response" | sed "s/${delimiter}.*//")
    response_status=$(echo "$response" | sed "s/.*${delimiter}//")
  else
    # If delimiter is not found, assume an error occurred
    echo "Curl Error: $response"
    exit 1
  fi

  # Validate if the response body is valid JSON
  if ! echo "$response_body" | jq . >/dev/null 2>&1; then
    echo "Error: Invalid JSON response in $method request to $url"
    echo "$response_body"
    exit 1
  fi

  # Output response body and status code on separate lines
  echo "$response_body"
  echo "$response_status"
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
# Perform the authentication curl request and capture response and status
AUTH_OUTPUT=$(perform_curl "POST" "$API_URL/authenticate" -H "Content-Type: application/json" -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

# Read the response body and status code
AUTH_RESPONSE=$(echo "$AUTH_OUTPUT" | head -n1)
AUTH_STATUS=$(echo "$AUTH_OUTPUT" | tail -n1)

echo "Auth Response: $AUTH_RESPONSE"  # Debugging line
echo "Auth Status: $AUTH_STATUS"      # Debugging line

# Check if authentication was successful
check_status $AUTH_STATUS "Authenticate User"

# Extract JWT token from the response
JWT_TOKEN=$(echo "$AUTH_RESPONSE" | jq -r '.token')

# Validate the JWT token
if [ -z "$JWT_TOKEN" ] || [ "$JWT_TOKEN" == "null" ]; then
  echo "Error: Failed to get JWT token"
  exit 1
fi

echo "Authentication Successful. JWT Token Obtained."

# 4. Create Test
echo "Creating Test..."
# Define the payload for creating the test
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

# Perform the create-test curl request and capture response and status
CREATE_TEST_OUTPUT=$(perform_curl "POST" "$API_URL/create-test" -H "Content-Type: application/json" -H "Authorization: Bearer $JWT_TOKEN" -d "$CREATE_TEST_PAYLOAD")

# Read the response body and status code
CREATE_TEST_RESPONSE=$(echo "$CREATE_TEST_OUTPUT" | head -n1)
CREATE_TEST_STATUS=$(echo "$CREATE_TEST_OUTPUT" | tail -n1)

echo "Create Test Response: $CREATE_TEST_RESPONSE"  # Debugging line
echo "Create Test Status: $CREATE_TEST_STATUS"      # Debugging line

# Check if creating the test was successful
if [ "$CREATE_TEST_STATUS" -ne 200 ] && [ "$CREATE_TEST_STATUS" -ne 201 ]; then
  echo "Error: Create Test failed with status $CREATE_TEST_STATUS"
  echo "Response: $CREATE_TEST_RESPONSE"  # Display the server's error message
  exit 1
fi

# Extract Test ID from the response
TEST_ID=$(echo "$CREATE_TEST_RESPONSE" | jq -r '.testID')

# Validate the Test ID
if [ -z "$TEST_ID" ] || [ "$TEST_ID" == "null" ]; then
  echo "Error: Failed to retrieve Test ID from Create Test response"
  exit 1
fi

echo "Test Created Successfully with Test ID: $TEST_ID"

# 5. Schedule Test
echo "Scheduling Test..."
# Define the schedule time (e.g., 10 seconds from now)
SCHEDULE_TIME=$(calculate_future_time 10)  # Schedule to start after 10 seconds

# Define the payload for scheduling the test
SCHEDULE_PAYLOAD=$(cat <<EOF
{
  "testID": "$TEST_ID",
  "schedule": "$SCHEDULE_TIME"
}
EOF
)

# Perform the schedule-test curl request and capture response and status
SCHEDULE_OUTPUT=$(perform_curl "POST" "$API_URL/schedule-test" -H "Content-Type: application/json" -H "Authorization: Bearer $JWT_TOKEN" -d "$SCHEDULE_PAYLOAD")

# Read the response body and status code
SCHEDULE_RESPONSE=$(echo "$SCHEDULE_OUTPUT" | head -n1)
SCHEDULE_STATUS=$(echo "$SCHEDULE_OUTPUT" | tail -n1)

echo "Schedule Test Response: $SCHEDULE_RESPONSE"  # Debugging line
echo "Schedule Test Status: $SCHEDULE_STATUS"      # Debugging line

# Check if scheduling was successful
if [ "$SCHEDULE_STATUS" -ne 200 ] && [ "$SCHEDULE_STATUS" -ne 201 ]; then
  echo "Error: Schedule Test failed with status $SCHEDULE_STATUS"
  echo "Response: $SCHEDULE_RESPONSE"  # Display the server's error message
  exit 1
fi

echo "Test Scheduled Successfully with status $SCHEDULE_STATUS"

# 6. Cancel Test
echo "Cancelling Test..."
# Define the payload for cancelling the test
CANCEL_PAYLOAD=$(cat <<EOF
{
  "testID": "$TEST_ID"
}
EOF
)

# Perform the cancel-test curl request and capture response and status
CANCEL_OUTPUT=$(perform_curl "POST" "$API_URL/cancel-test" -H "Content-Type: application/json" -H "Authorization: Bearer $JWT_TOKEN" -d "$CANCEL_PAYLOAD")

# Read the response body and status code
CANCEL_RESPONSE=$(echo "$CANCEL_OUTPUT" | head -n1)
CANCEL_STATUS=$(echo "$CANCEL_OUTPUT" | tail -n1)

echo "Cancel Test Response: $CANCEL_RESPONSE"  # Debugging line
echo "Cancel Test Status: $CANCEL_STATUS"      # Debugging line

# Check if cancelling was successful
check_status $CANCEL_STATUS "Cancel Test"
echo "Test Cancelled Successfully with status $CANCEL_STATUS"

# 7. Restart Test
echo "Restarting Test..."
# Define the payload for restarting the test
RESTART_PAYLOAD=$(cat <<EOF
{
  "testID": "$TEST_ID",
  "logRate": 1,
  "duration": 60
}
EOF
)

# Perform the restart-test curl request and capture response and status
RESTART_OUTPUT=$(curl -X POST "$API_URL/restart-test" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "$RESTART_PAYLOAD" \
  -s -w "__STATUS_DELIMITER__%{http_code}")

# Split the response body and status code
RESTART_RESPONSE=$(echo "$RESTART_OUTPUT" | sed "s/__STATUS_DELIMITER__.*//")
RESTART_STATUS=$(echo "$RESTART_OUTPUT" | sed "s/.*__STATUS_DELIMITER__//")

echo "Restart Test Response: $RESTART_RESPONSE"  # Debugging line
echo "Restart Test Status: $RESTART_STATUS"      # Debugging line

# Check if restarting was successful
if [ "$RESTART_STATUS" -ne 200 ] && [ "$RESTART_STATUS" -ne 201 ]; then
  echo "Error: Restart Test failed with status $RESTART_STATUS"
  echo "Response: $RESTART_RESPONSE"  # Display the server's error message
  exit 1
fi

echo "Test Restarted Successfully with status $RESTART_STATUS"

# 8. Save Results
echo "Saving Results..."
# Define the payload for saving results
SAVE_RESULTS_PAYLOAD=$(cat <<EOF
{
  "testID": "$TEST_ID",
  "completedAt": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}
EOF
)

# Perform the save-results curl request and capture response and status
SAVE_RESULTS_OUTPUT=$(perform_curl "POST" "$API_URL/save-results" -H "Content-Type: application/json" -H "Authorization: Bearer $JWT_TOKEN" -d "$SAVE_RESULTS_PAYLOAD")

# Read the response body and status code
SAVE_RESULTS_RESPONSE=$(echo "$SAVE_RESULTS_OUTPUT" | head -n1)
SAVE_RESULTS_STATUS=$(echo "$SAVE_RESULTS_OUTPUT" | tail -n1)

echo "Save Results Response: $SAVE_RESULTS_RESPONSE"  # Debugging line
echo "Save Results Status: $SAVE_RESULTS_STATUS"      # Debugging line

# Check if saving results was successful
if [ "$SAVE_RESULTS_STATUS" -ne 200 ] && [ "$SAVE_RESULTS_STATUS" -ne 201 ]; then
  echo "Error: Save Results failed with status $SAVE_RESULTS_STATUS"
  echo "Response: $SAVE_RESULTS_RESPONSE"  # Display the server's error message
  exit 1
fi

echo "Results Saved Successfully with status $SAVE_RESULTS_STATUS"

# 9. Get All Tests
echo "Getting All Tests..."
# Perform the get-all-tests curl request and capture response and status
GET_ALL_TESTS_OUTPUT=$(perform_curl "GET" "$API_URL/get-all-tests" -H "Authorization: Bearer $JWT_TOKEN")
GET_ALL_TESTS_RESPONSE=$(echo "$GET_ALL_TESTS_OUTPUT" | head -n1)
GET_ALL_TESTS_STATUS=$(echo "$GET_ALL_TESTS_OUTPUT" | tail -n1)

echo "Get All Tests Response: $GET_ALL_TESTS_RESPONSE"  # Debugging line
echo "Get All Tests Status: $GET_ALL_TESTS_STATUS"      # Debugging line

# Check if getting all tests was successful
check_status $GET_ALL_TESTS_STATUS "Get All Tests"
echo "Retrieved All Tests Successfully with status $GET_ALL_TESTS_STATUS"
echo "All Tests: $GET_ALL_TESTS_RESPONSE"

# 10. Get Test By ID
echo "Getting Test By ID..."
# Perform the get-test-by-id curl request and capture response and status
GET_TEST_BY_ID_OUTPUT=$(perform_curl "GET" "$API_URL/tests/$TEST_ID" -H "Authorization: Bearer $JWT_TOKEN")
GET_TEST_BY_ID_RESPONSE=$(echo "$GET_TEST_BY_ID_OUTPUT" | head -n1)
GET_TEST_BY_ID_STATUS=$(echo "$GET_TEST_BY_ID_OUTPUT" | tail -n1)

echo "Get Test By ID Response: $GET_TEST_BY_ID_RESPONSE"  # Debugging line
echo "Get Test By ID Status: $GET_TEST_BY_ID_STATUS"      # Debugging line

# Check if getting test by ID was successful
check_status $GET_TEST_BY_ID_STATUS "Get Test By ID"
echo "Retrieved Test By ID Successfully with status $GET_TEST_BY_ID_STATUS"
echo "Test Details: $GET_TEST_BY_ID_RESPONSE"

# 11. Metrics
echo "Getting Metrics..."
# Perform the metrics curl request and capture status
METRICS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/metrics")
check_status $METRICS_STATUS "Metrics"
echo "Metrics Retrieved Successfully with status $METRICS_STATUS"

echo "All tests passed successfully!"
