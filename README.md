А как?

Документация по API - BlessedApi.postman_collection.json

Запуск через make и docker:
1. Все закинуть на сервер в том виде, в котором это здесь находится
2. make migrate
3. make postgres api tgbot site

А что с бд?

Чтобы залить бекап:

docker exec -i postgres_db psql -U blessed blessed_db < blessed_db_backup.sql 