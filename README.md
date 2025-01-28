# user-task-reward-controler

запуск:
Запуcкаем Docker Desktop
собираем контейнеры
docker-compose up --build

Если контейнер с приложением не запустился, запускаем его повторно в докер десктопе

тестируем приложение черех postman

коллекция в корне приложения collection.json

перед тестированием приложения локально (не собирая приложение в контейнер), необходимо отредактировать путь к файлу миграции в пакете server в функции // runMigrations
изменить "file:///app/migration", на "file://migration",
далее развернуть базу данных в контенере db согласно инструкциям в docker-compose.yml касающихся сборки контейнера для базы данных

либо сконфигурировать базу локально согласно настройкам конфигурации в файле env
запустить приложение go run main.go




примеры curl -x для тестирования маршрутов в в комментариях к маршрутам файле router


маршрут auth/register
	/*
			curl -X POST "http://localhost:8080/auth/register" \
			-H "Content-Type: application/json" \
			-d '{
			"username": "john_doe",
				"password": "securepassword123",
				"email": "john.doe@example.com"
		}'
	*/
	//Пример успешного ответа
	/*
		{
		  "message": "User registered successfully",
		  "user_id": 123
		}
	*/
	

маршрут auth/login"

	/*
		curl -X POST "http://localhost:8080/auth/login" \
		-H "Content-Type: application/json" \
		-d '{
		  "username": "john_doe",
		  "password": "securepassword123"
		}'
	*/
	//Пример ответа от сервера

	/*
		{
		  "message": "Login successful",
		  "token": "your_jwt_token_here"
		}
	*/




маршрут api/task/create

	/*
			curl -X POST "http://localhost:8080/api/task/create" \
			-H "Content-Type: application/json" \
			-d '{
			"title": "New Task",
				"description": "This is a new task description.",
				"price": 50
		}'
	*/

маршрут api/task/all

	//curl -X GET "http://localhost:8080/api/task/all"
	router.HandleFunc("/task/all", handler.TaskGetAll).Methods("GET")
	/*
			curl -X POST "http://localhost:8080/api/task/123/complete" \
			-H "Content-Type: application/json" \
			-d '{
			"task_id": 456
		}'
	*/

маршрут api/users/123/refferer
	/*

		curl -X POST "http://localhost:8080/api/users/123/refferer" \
		-H "Content-Type: application/json" \
		-d '{
		  "refer_code": "ABC123"
		}'
	*/

маршрут	api/users/id/status"

	//curl -X GET "http://localhost:8080/api/users/123/status"

маршрут	api/users/leaderboard

	//curl -X GET "http://localhost:8080/api/users/leaderboard"

маршрут для поиска по имени или емейлу

	//примеры запросов
	//curl -X GET "http://localhost:8080/api/users/john_doe"
	//curl -X GET "http://localhost:8080/api/users/example@example.com"

	//пример ответа {"user_id": 123 }
	
}



Реализовать простой HTTP сервер для управления пользователями на языке Go


Основная логика приложения - создание пользователя, который выполняет какие-либо целевые действия, например, вводит реферальный код, подписывается на телеграм канал или твиттер и получает за это награду в виде поинтов. Награду за каждое задание вы можете определить самостоятельно, также вы можете добавить другие задачи, дайте волю фантазии


Нужно реализовать следующий функционал:
1. Middleware авторизация по Access token ко всем эндпоинтам (например JWT)
2. Реализация HTTP API:
* GET /users/{id}/status - вся доступная информация о пользователе
* GET /users/leaderboard - топ пользователей с самым большим балансом
* POST /users/{id}/task/complete - выполнение задания
* POST /users/{id}/referrer - ввод реферального кода (может быть id другого пользователя)
3. Создание хранилища для всех этих данных по пользователю (postgres). Обязательно использование инструментов для миграций (golang-migrate)
4. Сборка всего проекта в docker-compose


Дополнительные требования:
* Протестировать все указанные маршруты с помощью Postman или аналогичного инструмента
* Обеспечить обработку ошибок (например, неверные данные, несуществующие пользователи и т.д.)
* Не забывайте про принципы SOLID и чистоту кода


Если возникнут вопросы до или в процессе выполнения, не стесняйтесь задавать их нашей команде, мы с радостью поможем со всеми сложностями. Желаем удачи!
