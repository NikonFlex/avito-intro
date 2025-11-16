# avito-intro

В своем решении придерживался go-clean-template https://github.com/evrone/go-clean-template/tree/master

Выбор был сделан в сторону in-memory хранилища так как обьем пользователей очень мал и в базе данных нету смысла

В Makefile описано как собрать и запустить приложение без контейнера

при помощи docker compose build и docker compose up можно собрать и запустить приложение в контейнере