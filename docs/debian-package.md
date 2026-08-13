# Debian-пакет gwg

## Разделение ответственности

Debian-пакет выполняет только две задачи:

1. Устанавливает исполняемый файл `/usr/bin/gwg`, Bash completion и документацию.
2. Объявляет runtime-зависимости, которые APT устанавливает автоматически.

Пакет не создает пользователей или группы, не меняет владельца
`/etc/wireguard`, не создает WireGuard-интерфейс, не генерирует ключи, не
изменяет `sysctl`, не запускает службы и не вызывает `gwg init` из maintainer
scripts. Административные команды `gwg` выполняются через `sudo`.

Инициализация сервера выполняется пользователем отдельно:

```bash
sudo gwg init
```

Команда создает закрытые рабочие каталоги в `/etc/wireguard`, записывает
`/etc/sysctl.d/90-gwg.conf`, включает IPv4 forwarding и устанавливает сервер
`wg0` с сетью `10.0.0.1/24` и UDP-портом `51830`.

Каталог `/etc/wireguard` принадлежит `root:root` и имеет права `0700`. Поэтому
все команды `gwg`, которые читают или изменяют конфигурацию, управляют
WireGuard либо TC, необходимо запускать через `sudo`. Без `sudo` запускаются
только `gwg --help`, `gwg version` и `gwg completion`.

## Runtime-зависимости

Зависимости перечислены в `debian/control`:

| Пакет | Назначение |
| --- | --- |
| `bash-completion` | Автодополнение команд `gwg` в Bash (Recommends) |
| `wireguard-tools` | Команды `wg` и `wg-quick` |
| `iproute2` | Команды `ip` и `tc` |
| `iptables` | NAT и маршрутизация WireGuard |
| `procps` | Команда `sysctl` |
| `systemd` | Управление `wg-quick` и TC-службой |
| `sudo` | Выполнение системных команд текущей реализацией gwg |

Зависимости скачиваются только при установке пакета через APT. Программа
`gwg` не запускает `apt` или `dpkg`.

## Подготовка сборочной системы

Для сборки нужны `debhelper`, `dpkg-dev`, `fakeroot` и Go 1.26 или новее, как
указано в `src/go.mod` и `debian/control`:

```bash
sudo apt update
sudo apt install -y build-essential debhelper dpkg-dev fakeroot golang-go
go version
```

Если версия Go в используемом выпуске Debian ниже требуемой, необходимо
установить подходящий Go toolchain отдельно. Поскольку такой toolchain не
зарегистрирован в базе `dpkg`, после установки остальных build-зависимостей
проверка выполняется с отключенной проверкой `Build-Depends`:

```bash
dpkg-buildpackage -us -uc -b -d
```

Опция `-d` допустима только после ручной проверки `go version` и установки
остальных build-зависимостей. В окружении, где доступен Debian-пакет
`golang-any` версии 1.26 или новее, следует использовать обычную команду без
`-d`.

## Версии

При сборке используются два значения версии:

| Значение | Источник | Назначение |
| --- | --- | --- |
| Версия Debian-пакета | аргумент `version` или `debian/changelog` | Метаданные и имя файла `.deb` |
| Версия приложения | аргумент `version` или `debian/changelog` | Вывод команды `gwg version` |

Версия Debian-пакета задается первой записью `debian/changelog`:

```text
gwg (0.3.1-1) unstable; urgency=medium
```

Здесь `0.3.1` — версия программы, а `-1` — ревизия Debian-пакета. При обычной
сборке `debian/rules` берет `0.3.1` из changelog и встраивает ее в
`cmd.Version`.

Для нового выпуска сначала добавьте версию Debian-пакета в changelog:

```bash
dch -v 0.3.1-1 "Release 0.3.1"
```

Цель `deb` собирает пакет и передает версию в Go-бинарник:

```bash
make deb version=0.3.1
```

Значение передается в `debian/rules` через переменную окружения `VERSION`, а
затем встраивается в бинарник через Go `ldflags` и передается в
`dh_gencontrol`. Поэтому аргумент задает как версию приложения, так и версию и
имя `.deb`. Аргумент `version` обязателен, `debian/changelog` не изменяется.

## Сборка

Эквивалентная команда без Make, использующая версию приложения из changelog:

```bash
dpkg-buildpackage -us -uc -b
```

Основная команда сборки проекта требует явную версию Go-бинарника:

```bash
make deb version=0.3.1
```

Эта цель запускает `dpkg-buildpackage` с параметром `-d`, поскольку Go может
быть установлен отдельно и не зарегистрирован в `dpkg` как пакет `golang-any`.
Перед сборкой необходимо самостоятельно проверить `go version`; остальные
сборочные инструменты по-прежнему должны быть установлены.

Например, `make deb version=0.3.1` создаст `gwg_0.3.1_amd64.deb`, а установленная
из него команда `gwg version` покажет `0.3.1`. Для официального Debian-релиза
рекомендуется также добавить соответствующую запись в changelog.

Значение `dev` нельзя напрямую использовать как Debian-версию: поле `Version`
должно начинаться с цифры. Поэтому команда:

```bash
make deb version=dev
```

встраивает `dev` в Go-бинарник, но использует Debian-версию `0~dev` и создает
файл `gwg_0~dev_amd64.deb`.

Основной пакет после сборки копируется в каталог `vagrant/share`. Дефолтный путь
задается в начале `Makefile`:

```make
DEB_OUTPUT_DIR ?= vagrant/share
```

Путь можно переопределить без изменения `Makefile`:

```bash
make deb version=0.3.1 DEB_OUTPUT_DIR=dist
```

Перед созданием пакета `debian/rules` переходит в `src` и выполняет:

```bash
cd src
go test ./...
go build -trimpath -buildmode=pie ...
```

Готовые файлы создаются в родительском каталоге, например:

```text
../gwg_0.3.1-1_amd64.deb
../gwg-dbgsym_0.3.1-1_amd64.deb
../gwg_0.3.1-1_amd64.changes
```

## Проверка содержимого

```bash
dpkg-deb --info ../gwg_0.3.1-1_amd64.deb
dpkg-deb --contents ../gwg_0.3.1-1_amd64.deb
lintian ../gwg_0.3.1-1_amd64.changes
```

В основном пакете должны находиться `/usr/bin/gwg`, man-страница, документация
и `/usr/share/bash-completion/completions/gwg`. Пакет не должен содержать
maintainer script, создающий пользователя или меняющий владельца
`/etc/wireguard`.

## Установка

Для локального `.deb` следует использовать APT, а не `dpkg -i`, поскольку APT
разрешает и скачивает зависимости:

```bash
sudo apt install ./gwg_0.3.1-1_amd64.deb
```

После установки можно проверить зависимости и версию:

```bash
dpkg-query -W gwg bash-completion wireguard-tools iproute2 iptables procps systemd sudo
stat -c '%U:%G %a %n' /etc/wireguard
gwg version
```

После открытия новой Bash-сессии можно проверить completion:

```bash
type _gwg
sudo gwg server <TAB><TAB>
```

Ожидаемые владелец и права каталога:

```text
root:root 700 /etc/wireguard
```

Сервер на этом этапе еще не настроен.

## Инициализация сервера

Для стандартной конфигурации:

```bash
sudo gwg init
```

Для явного задания параметров вместо `init`:

```bash
sudo gwg server install --name wg0 --network 10.0.0.1/24 --port 51830
```

После инициализации:

```bash
sudo systemctl status wg-quick@wg0.service
sudo wg show wg0
```

## Обновление и удаление

Обновление заменяет бинарник, но не запускает инициализацию и не изменяет
файлы в `/etc/wireguard`:

```bash
sudo apt install ./gwg_NEW_VERSION_amd64.deb
```

Удаление пакета:

```bash
sudo apt remove gwg
```

Поскольку конфигурации создаются самой программой, а не пакетом, APT их не
удаляет. Удалять WireGuard-конфигурацию и ключи следует только отдельной
осознанной административной операцией.

## Проверка в Vagrant

Скопируйте пакет в общий каталог и установите его на серверной VM:

```bash
cp ../gwg_0.3.1-1_amd64.deb vagrant/share/
vagrant up server
vagrant ssh server
sudo apt install /home/vagrant/share/gwg_0.3.1-1_amd64.deb
sudo gwg init
```

При проверке установки VM должна иметь доступ к Debian-репозиториям для
скачивания объявленных зависимостей.
