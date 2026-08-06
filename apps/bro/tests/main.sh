url="http://127.0.0.1:8000"
acc="Accept: application/json"

# --- Testing Framework ---
assert_equals() {
    local expected="$1"
    local actual="$2"
    local test_name="$3"

    if [ "$expected" == "$actual" ]; then
        echo "✅ PASS: $test_name"
    else
        echo "❌ FAIL: $test_name (Expected '$expected', but got '$actual')"
        exit 1
    fi
}

assert_not_empty() {
    local expected="$1"
    local actual="$2"
    local test_name="$3"

    if [ -z "$actual" ]; then
        printf "✅ PASS: %s\n%s" "$test_name" "$actual"
    else
        echo "❌ FAIL: $test_name (Expected '$expected', but got '$actual')"
        exit 1
    fi
}

# --- Test Cases ---
test_index() {
    act=$(curl -X GET --location "$url" -H "$acc")
    exp='{"message":"🤖"}'
    assert_equals "$exp" "$act" "test_index"
}

test_news() {
  bot=1
  query="situation in iran"
  since="2025-01-01T05:00:00Z"
  act=$(curl -X PUT --location "$url/news/$bot/query/$query/since/$since" -H "$acc")

}



# --- Run Tests ---
echo "Running tests..."
test_index
echo "All tests passed successfully!"