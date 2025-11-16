# Avito-intership_back
**Тестовое задание: распределение PR-ревьюеров**
**GO + Gin + GORM + PostgreSQL + Docker-compose**

## Описание 

Сервис реализует систему распределения PR-reviewer в команде: 

-- Создание команды разработчиков
-- Автоматической назначение до 2 свободных разработчиков на открытый Pull Request 
-- Перераспределение разработчиков с сохранением количества 
-- Получение статистики по командам: Сколько открытых PR числится на разработчике

## Запуск

``bash 

sudo docker compose down -v
sudo docker-compose up --build

## Тестирование 

Для нагрузочного тестирования использовал консольную утилиту 'hey'

при Тестировании GET http://localhost:8080/health

 RPS -- 1000
 SLI -- 100мс 

 Summary:
  Total:	0.6485 secs
  Slowest:	0.0536 secs
  Fastest:	0.0003 secs
  Average:	0.0062 secs
  Requests/sec:	15419.5638

При тестировании Get "http://localhost:8080/team/get?team_name=backend"
 

 RPS -- 1000
 SLI -- 100мс 

 Summary:
  Total:	1.8034 secs
  Slowest:	1.0482 secs
  Fastest:	0.0015 secs
  Average:	0.1398 secs
  Requests/sec:	554.5106

При тестировании POST   http://localhost:8080/pullRequest/create


 RPS -- 1000
 SLI -- 50мс 
   Total:	1.4123 secs
  Slowest:	0.3886 secs
  Fastest:	0.0011 secs
  Average:	0.0598 secs
  Requests/sec:	708.0523
