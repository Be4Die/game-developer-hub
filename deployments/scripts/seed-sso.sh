#!/usr/bin/env bash
# seed-sso — заполнить SSO тестовыми пользователями, верифицировать email и установить роли.
# Использование: ./seed-sso.sh [output-csv]
set -euo pipefail

OUTPUT_CSV="${1:-./docs/test-users.csv}"
GATEWAY="${GATEWAY:-http://localhost:8080}"

echo "=== Сидирование SSO ==="
echo "CSV: $OUTPUT_CSV"
echo

# Создаём CSV с заголовком
mkdir -p "$(dirname "$OUTPUT_CSV")"
echo "email,password,display_name,role" > "$OUTPUT_CSV"

# Фиксированные тестовые пользователи (email|password|display_name|role)
# role: 1=developer, 2=moderator, 3=admin
USERS_LIST=(
  "developer@test.local|Developer123!dev|Разработчик|1"
  "moderator@test.local|Moderator123!mod|Модератор|2"
  "admin@test.local|Admin123!admin|Администратор|3"
  "user1@test.local|User1Password!|Пользователь 1|1"
  "user2@test.local|User2Password!|Пользователь 2|1"
  "user3@test.local|User3Password!|Пользователь 3|1"
)

ROLE_NAMES=( "" "developer" "moderator" "admin" )

for entry in "${USERS_LIST[@]}"; do
  IFS='|' read -r email password display_name role <<< "$entry"
  role_name="${ROLE_NAMES[$role]}"

  echo "Регистрация: $email"
  resp=$(curl -s -X POST "$GATEWAY/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"password\":\"$password\",\"display_name\":\"$display_name\"}")

  # Извлекаем user_id из ответа (JSON: {"user":{"id":"..."}})
  user_id=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | sed 's/"id":"//;s/"//')

  # Получаем код верификации из Valkey
  code=$(docker exec gdh-sso-valkey valkey-cli GET "sso:verify:$email" 2>/dev/null || echo "")

  if [ -n "$code" ] && [ "$code" != "(nil)" ]; then
    echo "  Код верификации: $code"
    verify_resp=$(curl -s -X POST "$GATEWAY/api/v1/auth/verify-email" \
      -H "Content-Type: application/json" \
      -d "{\"verification_code\":\"$code\"}")

    if echo "$verify_resp" | grep -q '"success":true'; then
      echo "  Email подтверждён"
    else
      echo "  Ошибка верификации: $verify_resp"
    fi
  else
    echo "  Код не найден в Valkey"
  fi

  # Устанавливаем роль через БД (ChangeRole ещё не реализован в сервисе)
  if [ -n "$user_id" ] && [ "$role" != "1" ]; then
    echo "  Установка роли: $role_name (user_id=$user_id)"
    docker exec gdh-sso-postgres psql -U postgres -d sso \
      -c "UPDATE users SET role=$role WHERE id='$user_id';" >/dev/null 2>&1
    echo "  Роль установлена"
  fi

  # Записываем в CSV
  echo "$email,$password,$display_name,$role_name" >> "$OUTPUT_CSV"
done

echo
echo "=== CSV файл ==="
cat "$OUTPUT_CSV"
echo
echo "=== Готово ==="
