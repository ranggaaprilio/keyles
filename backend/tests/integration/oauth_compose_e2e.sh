#!/usr/bin/env bash
set -euo pipefail

API_URL="${KEYLES_E2E_API_URL:-http://localhost:8080}"
FRONTEND_URL="${KEYLES_E2E_FRONTEND_URL:-http://localhost:3000}"
CALLBACK_URL="${KEYLES_E2E_CALLBACK_URL:-http://localhost:9999/callback}"
CONTAINER_ENGINE="${CONTAINER_ENGINE:-podman}"
USER_ID="00000000-0000-0000-0000-000000000062"
PUBLIC_CLIENT="dev_public_client"
SECOND_CLIENT="dev_second_client"
CONFIDENTIAL_CLIENT="dev_confidential_client"
CONFIDENTIAL_SECRET="dev_client_secret_change_in_production"
USER_EMAIL="user@dev-tenant.com"
USER_PASSWORD="user123"
SPOOFED_IP="203.0.113.99"
VERIFIER="compose_e2e_verifier_abcdefghijklmnopqrstuvwxyz0123456789ABCDE"
CHALLENGE="$(printf '%s' "$VERIFIER" | openssl dgst -sha256 -binary | openssl base64 -A | tr '+/' '-_' | tr -d '=')"
TMP_DIR="$(mktemp -d)"
COOKIE_JAR="$TMP_DIR/cookies.txt"
NO_SESSION_COOKIE_JAR="$TMP_DIR/no-session-cookies.txt"
REDIS_STOPPED=false

cleanup() {
  if [ "$REDIS_STOPPED" = true ]; then
    "$CONTAINER_ENGINE" start keyles-redis >/dev/null || true
  fi
  "$CONTAINER_ENGINE" exec keyles-postgres psql -q -U keyles -d keyles \
    -c "update users set status='active' where id='$USER_ID';" >/dev/null || true
  "$CONTAINER_ENGINE" exec keyles-postgres psql -q -U keyles -d keyles \
    -c "update user_role_assignments set is_active=true where user_id='$USER_ID';" >/dev/null || true
  "$CONTAINER_ENGINE" exec keyles-redis redis-cli FLUSHDB >/dev/null || true
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

fail() {
  printf 'compose e2e failure: %s\n' "$*" >&2
  exit 1
}

assert_eq() {
  [ "$1" = "$2" ] || fail "expected '$2', got '$1'"
}

assert_contains() {
  case "$1" in
    *"$2"*) ;;
    *) fail "expected '$1' to contain '$2'" ;;
  esac
}

redis() {
  "$CONTAINER_ENGINE" exec keyles-redis redis-cli "$@"
}

sql() {
  "$CONTAINER_ENGINE" exec keyles-postgres psql -At -U keyles -d keyles -c "$1"
}

authorize() {
  local client_id="$1"
  local state="$2"
  local cookie_jar="${3:-$COOKIE_JAR}"
  local prompt="${4:-}"
  local max_age="${5:-}"
  local redirect_uri="${6:-$CALLBACK_URL}"
  local headers="$TMP_DIR/auth-headers.txt"
  local args=(
    --get "$API_URL/oauth2/auth"
    --data-urlencode "client_id=$client_id"
    --data-urlencode "redirect_uri=$redirect_uri"
    --data-urlencode "response_type=code"
    --data-urlencode "scope=openid profile email offline_access"
    --data-urlencode "state=$state"
    --data-urlencode "code_challenge=$CHALLENGE"
    --data-urlencode "code_challenge_method=S256"
  )
  if [ -n "$prompt" ]; then
    args+=(--data-urlencode "prompt=$prompt")
  fi
  if [ -n "$max_age" ]; then
    args+=(--data-urlencode "max_age=$max_age")
  fi
  curl -sS -D "$headers" -o /dev/null -b "$cookie_jar" \
    -H "X-Forwarded-For: $SPOOFED_IP" "${args[@]}"
  tr -d '\r' < "$headers" | sed -n 's/^Location: //p'
}

transaction_id() {
  printf '%s' "$1" | sed -n 's/.*transaction_id=\([^&]*\).*/\1/p'
}

login() {
  local transaction="$1"
  local password="${2:-$USER_PASSWORD}"
  local cookie_jar="${3:-$COOKIE_JAR}"
  RESPONSE_HEADERS="$TMP_DIR/login-headers.txt"
  RESPONSE_BODY="$TMP_DIR/login-body.json"
  RESPONSE_STATUS="$(curl -sS -D "$RESPONSE_HEADERS" -o "$RESPONSE_BODY" -w '%{http_code}' \
    -c "$cookie_jar" -H 'Content-Type: application/json' \
    -H "X-Forwarded-For: $SPOOFED_IP" \
    -d "{\"transaction_id\":\"$transaction\",\"email\":\"$USER_EMAIL\",\"password\":\"$password\"}" \
    "$API_URL/oauth2/login")"
}

consent_details() {
  local transaction="$1"
  local cookie_jar="${2:-$COOKIE_JAR}"
  RESPONSE_BODY="$TMP_DIR/consent-details.json"
  RESPONSE_STATUS="$(curl -sS -o "$RESPONSE_BODY" -w '%{http_code}' -b "$cookie_jar" \
    "$API_URL/oauth2/consent/$transaction")"
}

consent() {
  local transaction="$1"
  local csrf="$2"
  local approved="$3"
  local cookie_jar="${4:-$COOKIE_JAR}"
  RESPONSE_BODY="$TMP_DIR/consent.json"
  RESPONSE_STATUS="$(curl -sS -o "$RESPONSE_BODY" -w '%{http_code}' -b "$cookie_jar" \
    -H 'Content-Type: application/json' \
    -d "{\"transaction_id\":\"$transaction\",\"interaction_csrf_token\":\"$csrf\",\"approved\":$approved}" \
    "$API_URL/oauth2/consent")"
}

approve() {
  local transaction="$1"
  local cookie_jar="${2:-$COOKIE_JAR}"
  consent_details "$transaction" "$cookie_jar"
  assert_eq "$RESPONSE_STATUS" "200"
  local csrf
  csrf="$(jq -r '.interaction_csrf_token' < "$RESPONSE_BODY")"
  consent "$transaction" "$csrf" true "$cookie_jar"
  assert_eq "$RESPONSE_STATUS" "200"
  REDIRECT_URL="$(jq -r '.redirect_url' < "$RESPONSE_BODY")"
}

token() {
  local code="$1"
  local verifier="$2"
  local client_id="${3:-$PUBLIC_CLIENT}"
  local client_secret="${4:-}"
  local args=(
    -sS -o "$TMP_DIR/token.json" -w '%{http_code}'
    -H 'Content-Type: application/x-www-form-urlencoded'
    --data-urlencode 'grant_type=authorization_code'
    --data-urlencode "client_id=$client_id"
    --data-urlencode "code=$code"
    --data-urlencode "redirect_uri=$CALLBACK_URL"
    --data-urlencode "code_verifier=$verifier"
  )
  if [ -n "$client_secret" ]; then
    args+=(--data-urlencode "client_secret=$client_secret")
  fi
  RESPONSE_STATUS="$(curl "${args[@]}" "$API_URL/oauth2/token")"
  RESPONSE_BODY="$TMP_DIR/token.json"
}

printf 'Checking live Podman/Compose services...\n'
curl -fsS "$API_URL/health" >/dev/null
curl -fsS "$FRONTEND_URL/" >/dev/null
assert_eq "$(redis PING)" "PONG"
redis FLUSHDB >/dev/null

printf 'Checking discovery, JWKS, and CORS...\n'
assert_eq "$(curl -fsS "$API_URL/.well-known/openid-configuration" | jq -r '.issuer')" "$API_URL"
[ "$(curl -fsS "$API_URL/.well-known/jwks.json" | jq '.keys | length')" -gt 0 ] || fail "JWKS is empty"
CORS_HEADERS="$(curl -sS -D - -o /dev/null -X OPTIONS "$API_URL/oauth2/login" \
  -H "Origin: $FRONTEND_URL" -H 'Access-Control-Request-Method: POST' | tr -d '\r')"
assert_contains "$CORS_HEADERS" "Access-Control-Allow-Origin: $FRONTEND_URL"
assert_contains "$CORS_HEADERS" "Access-Control-Allow-Credentials: true"

printf 'Checking approve, replay, token, userinfo, and refresh...\n'
LOCATION="$(authorize "$PUBLIC_CLIENT" compose-approve)"
assert_contains "$LOCATION" "$FRONTEND_URL/oauth2/login"
TRANSACTION="$(transaction_id "$LOCATION")"
login "$TRANSACTION"
assert_eq "$RESPONSE_STATUS" "200"
COOKIE_HEADER="$(tr -d '\r' < "$RESPONSE_HEADERS" | sed -n 's/^Set-Cookie: //p')"
assert_contains "$COOKIE_HEADER" "HttpOnly"
assert_contains "$COOKIE_HEADER" "SameSite=Lax"
case "$COOKIE_HEADER" in *"Domain="*) fail "SSO cookie must remain host-only" ;; esac
approve "$TRANSACTION"
assert_contains "$REDIRECT_URL" "state=compose-approve"
CODE="$(printf '%s' "$REDIRECT_URL" | sed -n 's/.*[?&]code=\([^&]*\).*/\1/p')"
[ -n "$CODE" ] || fail "approve callback did not include a code"
consent "$TRANSACTION" "$(jq -r '.interaction_csrf_token' < "$TMP_DIR/consent-details.json")" true
assert_eq "$RESPONSE_STATUS" "410"
token "$CODE" "$VERIFIER"
assert_eq "$RESPONSE_STATUS" "200"
ACCESS_TOKEN="$(jq -r '.access_token' < "$RESPONSE_BODY")"
REFRESH_TOKEN="$(jq -r '.refresh_token' < "$RESPONSE_BODY")"
[ -n "$ACCESS_TOKEN" ] && [ -n "$REFRESH_TOKEN" ] || fail "token response is incomplete"
assert_eq "$(curl -sS -o "$TMP_DIR/userinfo.json" -w '%{http_code}' \
  -H "Authorization: Bearer $ACCESS_TOKEN" "$API_URL/oauth2/userinfo")" "200"
assert_eq "$(jq -r '.sub' < "$TMP_DIR/userinfo.json")" "$USER_ID"
assert_eq "$(curl -sS -o "$TMP_DIR/refresh.json" -w '%{http_code}' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=refresh_token' --data-urlencode "client_id=$PUBLIC_CLIENT" \
  --data-urlencode "refresh_token=$REFRESH_TOKEN" "$API_URL/oauth2/token")" "200"

printf 'Checking deny, session reuse, prompt, max_age, and silent errors...\n'
LOCATION="$(authorize "$SECOND_CLIENT" compose-reuse)"
assert_contains "$LOCATION" "$FRONTEND_URL/oauth2/consent"
TRANSACTION="$(transaction_id "$LOCATION")"
consent_details "$TRANSACTION"
CSRF="$(jq -r '.interaction_csrf_token' < "$RESPONSE_BODY")"
consent "$TRANSACTION" "$CSRF" false
assert_eq "$RESPONSE_STATUS" "200"
assert_contains "$(jq -r '.redirect_url' < "$RESPONSE_BODY")" "error=access_denied"
assert_contains "$(authorize "$SECOND_CLIENT" compose-prompt "$COOKIE_JAR" login)" "$FRONTEND_URL/oauth2/login"
assert_contains "$(authorize "$SECOND_CLIENT" compose-age "$COOKIE_JAR" '' 0)" "$FRONTEND_URL/oauth2/login"
assert_contains "$(authorize "$PUBLIC_CLIENT" compose-none "$NO_SESSION_COOKIE_JAR" none)" "error=login_required"
assert_contains "$(authorize "$PUBLIC_CLIENT" compose-none-session "$COOKIE_JAR" none)" "error=consent_required"

printf 'Checking disabled-user and removed-role forced login...\n'
sql "update users set status='disabled' where id='$USER_ID';" >/dev/null
assert_contains "$(authorize "$SECOND_CLIENT" compose-disabled)" "$FRONTEND_URL/oauth2/login"
sql "update users set status='active' where id='$USER_ID';" >/dev/null
sql "update user_role_assignments set is_active=false where user_id='$USER_ID' and client_id='$SECOND_CLIENT';" >/dev/null
assert_contains "$(authorize "$SECOND_CLIENT" compose-role-removed)" "$FRONTEND_URL/oauth2/login"
sql "update user_role_assignments set is_active=true where user_id='$USER_ID' and client_id='$SECOND_CLIENT';" >/dev/null

printf 'Checking invalid callback and PKCE rejection...\n'
assert_contains "$(authorize "$PUBLIC_CLIENT" compose-invalid "$COOKIE_JAR" '' '' 'http://attacker.example/callback')" "$FRONTEND_URL/oauth2/error?error=invalid_redirect_uri"
LOCATION="$(authorize "$PUBLIC_CLIENT" compose-wrong-pkce)"
TRANSACTION="$(transaction_id "$LOCATION")"
approve "$TRANSACTION"
CODE="$(printf '%s' "$REDIRECT_URL" | sed -n 's/.*[?&]code=\([^&]*\).*/\1/p')"
token "$CODE" "wrong-verifier"
assert_eq "$RESPONSE_STATUS" "400"
assert_eq "$(jq -r '.error' < "$RESPONSE_BODY")" "invalid_grant"

printf 'Checking confidential token, introspection, and revocation...\n'
LOCATION="$(authorize "$CONFIDENTIAL_CLIENT" compose-confidential)"
TRANSACTION="$(transaction_id "$LOCATION")"
approve "$TRANSACTION"
CODE="$(printf '%s' "$REDIRECT_URL" | sed -n 's/.*[?&]code=\([^&]*\).*/\1/p')"
token "$CODE" "$VERIFIER" "$CONFIDENTIAL_CLIENT" "$CONFIDENTIAL_SECRET"
assert_eq "$RESPONSE_STATUS" "200"
CONF_ACCESS_TOKEN="$(jq -r '.access_token' < "$RESPONSE_BODY")"
CONF_REFRESH_TOKEN="$(jq -r '.refresh_token' < "$RESPONSE_BODY")"
assert_eq "$(curl -sS -o "$TMP_DIR/introspect.json" -w '%{http_code}' \
  -u "$CONFIDENTIAL_CLIENT:$CONFIDENTIAL_SECRET" \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode "token=$CONF_ACCESS_TOKEN" "$API_URL/oauth2/introspect")" "200"
assert_eq "$(jq -r '.active' < "$TMP_DIR/introspect.json")" "true"
assert_eq "$(curl -sS -o /dev/null -w '%{http_code}' \
  -u "$CONFIDENTIAL_CLIENT:$CONFIDENTIAL_SECRET" \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode "token=$CONF_REFRESH_TOKEN" --data-urlencode 'token_type_hint=refresh_token' \
  "$API_URL/oauth2/revoke")" "200"

printf 'Checking fixed-window throttling and direct-peer audit IP...\n'
redis FLUSHDB >/dev/null
LOCATION="$(authorize "$PUBLIC_CLIENT" compose-throttle "$NO_SESSION_COOKIE_JAR")"
TRANSACTION="$(transaction_id "$LOCATION")"
for attempt in 1 2 3 4 5; do
  login "$TRANSACTION" wrong-password "$NO_SESSION_COOKIE_JAR"
  assert_eq "$RESPONSE_STATUS" "401"
done
login "$TRANSACTION" wrong-password "$NO_SESSION_COOKIE_JAR"
assert_eq "$RESPONSE_STATUS" "429"
AUDIT_IP="$(sql "select coalesce(ip_address::text,'') from audit_logs where event_type='oauth_login_throttled' order by created_at desc limit 1;")"
[ -n "$AUDIT_IP" ] || fail "missing throttled audit event"
[ "$AUDIT_IP" != "$SPOOFED_IP" ] || fail "audit trusted spoofed forwarded IP"

printf 'Checking logout and Redis outage fail-closed behavior...\n'
redis FLUSHDB >/dev/null
LOCATION="$(authorize "$PUBLIC_CLIENT" compose-logout "$NO_SESSION_COOKIE_JAR")"
TRANSACTION="$(transaction_id "$LOCATION")"
login "$TRANSACTION" "$USER_PASSWORD" "$COOKIE_JAR"
assert_eq "$RESPONSE_STATUS" "200"
LOCATION="$(authorize "$PUBLIC_CLIENT" compose-outage-consent "$COOKIE_JAR")"
CONSENT_TRANSACTION="$(transaction_id "$LOCATION")"
LOCATION="$(authorize "$PUBLIC_CLIENT" compose-outage-login "$NO_SESSION_COOKIE_JAR")"
LOGIN_TRANSACTION="$(transaction_id "$LOCATION")"
"$CONTAINER_ENGINE" stop keyles-redis >/dev/null
REDIS_STOPPED=true
assert_contains "$(authorize "$PUBLIC_CLIENT" compose-outage "$NO_SESSION_COOKIE_JAR")" "error=temporarily_unavailable"
login "$LOGIN_TRANSACTION" "$USER_PASSWORD" "$NO_SESSION_COOKIE_JAR"
assert_eq "$RESPONSE_STATUS" "503"
consent_details "$CONSENT_TRANSACTION" "$COOKIE_JAR"
assert_eq "$RESPONSE_STATUS" "503"
LOGOUT_HEADERS="$TMP_DIR/logout-headers.txt"
assert_eq "$(curl -sS -D "$LOGOUT_HEADERS" -o /dev/null -w '%{http_code}' -b "$COOKIE_JAR" \
  -X POST "$API_URL/oauth2/logout")" "204"
assert_contains "$(tr -d '\r' < "$LOGOUT_HEADERS")" "Max-Age=0"
"$CONTAINER_ENGINE" start keyles-redis >/dev/null
REDIS_STOPPED=false
for _ in $(seq 1 20); do
  if [ "$(redis PING 2>/dev/null || true)" = "PONG" ]; then
    break
  fi
  sleep 1
done
assert_eq "$(redis PING)" "PONG"

printf 'Compose OAuth E2E matrix passed.\n'
