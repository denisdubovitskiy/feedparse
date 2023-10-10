# Запуск:
# make splash-start - запускаем браузер
# make update-config - обновляем конфиг и инициализируем базу
# make generate - генерируем гошный код
# make run - запускаем парсер

bin-deps:
	@ls $(CURDIR)/bin/sqlc &> /dev/null || GOBIN=$(CURDIR)/bin go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@ls $(CURDIR)/bin/goimports &> /dev/null || GOBIN=$(CURDIR)/bin go install golang.org/x/tools/cmd/goimports@latest

generate: bin-deps
	PATH=$(CURDIR)/bin sqlc generate

goimports:
	PATH=$(CURDIR)/bin goimports -w $(CURDIR)/

COMPOSE := docker-compose \
	--file $(CURDIR)/docker-compose.yml \
	--project-name feedparser

ps:
	$(COMPOSE) ps

splash-start:
	$(COMPOSE) up --detach splash

splash-stop:
	$(COMPOSE) stop splash

splash-restart: splash-stop splash-start

splash-rm: splash-stop
	$(COMPOSE) rm --force splash

update-config:
	go run $(CURDIR)/cmd/config/main.go \
		-config $(CURDIR)/config.yml \
		-database $(CURDIR)/database.sqlite3

run:
	go run $(CURDIR)/cmd/parser/main.go \
		-database $(CURDIR)/database.sqlite3

clean:
	$(RM) $(CURDIR)/database.sqlite3

build:
	go build $(CURDIR)/...

precommit: goimports build

runall: splash-start update-config generate run

.PHONY: bin-deps generate splash-start splash-stop \
	splash-restart splash-rm update-config run clean \
	goimports precommit runall
