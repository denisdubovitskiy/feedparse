# Запуск:
# make chrome-start - запускаем браузер
# make update-config - обновляем конфиг и инициализируем базу
# make generate - генерируем гошный код
# make run - запускаем парсер

ifneq (,$(wildcard ./.env))
	include .env
	export
endif

bin-deps:
	@ls $(CURDIR)/bin/sqlc &> /dev/null || GOBIN=$(CURDIR)/bin go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@ls $(CURDIR)/bin/goimports &> /dev/null || GOBIN=$(CURDIR)/bin go install golang.org/x/tools/cmd/goimports@latest

generate: bin-deps
	PATH=$(CURDIR)/bin sqlc generate
	PATH=$(CURDIR)/bin:$$PATH go generate $(CURDIR)/...

goimports:
	PATH=$(CURDIR)/bin goimports -w $(CURDIR)/


COMPOSE := docker compose \
	--file $(CURDIR)/docker-compose.yml \
	--project-name feedparser

DATABASE := $(CURDIR)/storage/database.sqlite3

ps:
	$(COMPOSE) ps

chrome-start:
	$(COMPOSE) up --detach chrome

chrome-stop:
	$(COMPOSE) stop chrome

chrome-restart: chrome-stop chrome-start

chrome-rm: chrome-stop
	$(COMPOSE) rm --force chrome

app-start:
	$(COMPOSE) up --detach app

app-stop:
	$(COMPOSE) stop app

app-restart: app-stop app-start

app-rm: app-stop
	$(COMPOSE) rm --force app

update-config:
	go run $(CURDIR)/cmd/config/main.go \
		-config $(CURDIR)/config/config.yml \
		-database $(DATABASE)

run:
	go run $(CURDIR)/cmd/parser/main.go \
		-database $(DATABASE)

clean:
	$(RM) $(DATABASE)

build:
	go build $(CURDIR)/...

build-parser:
	go build -o $(CURDIR)/bin/parser $(CURDIR)/cmd/parser

build-image:
	$(COMPOSE) build --no-cache app

precommit: goimports build

runall: chrome-start update-config generate run

.PHONY: bin-deps generate chrome-start chrome-stop \
	chrome-restart chrome-rm update-config run clean \
	goimports precommit runall build-parser build-image
