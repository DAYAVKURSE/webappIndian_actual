#!/bin/bash
# Этот скрипт обновляет все fetch и WebSocket запросы с жестко заданным протоколом на использование переменных из config.js

# Обновление импортов
find site/blessed/src -type f -name "*.js*" -exec sed -i 's/import { API_BASE_URL } from/import { API_BASE_URL, API_PROTOCOL } from/g' {} \;
find site/blessed/src -type f -name "*.js*" -exec sed -i 's/import { API_BASE_URL,/import { API_BASE_URL, API_PROTOCOL,/g' {} \;

# Обновление запросов fetch с HTTPS на API_PROTOCOL
find site/blessed/src -type f -name "*.js*" -exec sed -i 's/fetch(`https:\/\/${API_BASE_URL}/fetch(`${API_PROTOCOL}:\/\/${API_BASE_URL}/g' {} \;

# Обновление импортов для WebSocket
find site/blessed/src -type f -name "*.js*" -exec sed -i 's/import { API_BASE_URL } from/import { API_BASE_URL, WS_PROTOCOL } from/g' {} \;
find site/blessed/src -type f -name "*.js*" -exec sed -i 's/import { API_BASE_URL,/import { API_BASE_URL, WS_PROTOCOL,/g' {} \;

# Обновление WebSocket с WSS на WS_PROTOCOL
find site/blessed/src -type f -name "*.js*" -exec sed -i 's/WebSocket(`wss:\/\/${API_BASE_URL}/WebSocket(`${WS_PROTOCOL}:\/\/${API_BASE_URL}/g' {} \;

echo "Обновление завершено. Проверьте файлы на корректность изменений." 