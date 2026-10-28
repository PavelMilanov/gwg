# gwu — менеджер WireGuard-сервера

Консольная утилита для Linux: настройка WireGuard, управление пользователями,
локальными шаблонами конфигурации и общим ограничением скорости через Netlink.
Репозиторий выпускает один бинарник и Debian-пакет `gwu`. Лицензия — MIT.

## Возможности

- Первоначальная настройка IPv4 WireGuard-сервера.
- Добавление, удаление, блокировка и разблокировка клиентов.
- Создание клиентских конфигураций и настройка шаблонов.
- Один общий egress-канал HTB/FQ-CoDel на WireGuard-интерфейсе.
- Bash completion и пакетирование для Debian.

QoS ограничивает суммарный download клиентов при отправке через VPN-интерфейс.
Персональных лимитов, upload shaping и ACL нет. Состояние WireGuard смотрят
штатной командой `sudo wg show`.

**Команды изменения пользователей регенерируют конфигурацию и перезапускают
WireGuard-интерфейс. Это может прервать соединения.**

## Требования и установка

Целевая система — Debian 13 с systemd, WireGuard, nftables и iproute2.
Скачайте пакет со страницы [релизов](https://github.com/PavelMilanov/gwg/releases)
при наличии подходящего выпуска или соберите его самостоятельно:

```bash
make deb
sudo apt install ./vagrant/share/gwu_0.3.0_amd64.deb
gwu version
gwu --help
```

APT устанавливает зависимости. Сам пакет не создаёт VPN, не включает службы
и не переносит данные предыдущей версии gwg. Для неё сначала прочитайте
[совместимость и переход](docs/migration.md).

До первоначальной настройки обеспечьте доступность UDP-порта сервера и
разрешите нужный forwarded-трафик своим firewall. Утилита включает IPv4 forwarding
при `wireguard init` и создаёт NAT hooks, но не управляет filter policy, DNS-сервером
или специальной маршрутизацией. Детали — в [руководстве](docs/usage.md).

На новом сервере без существующего wg0:

```bash
sudo gwu wireguard init
sudo wg show wg0
sudo gwu wireguard user add alice
sudo gwu wireguard user list
```

Клиентский конфиг создаётся в
`/etc/wireguard/gwu/wireguard/configs/alice.conf`. Он содержит private key:
передавайте его клиенту по защищённому каналу.

## Сборка и проверки

Установите Go версии из [src/go.mod](src/go.mod). Из корня репозитория:

```bash
make test
(cd src && go vet ./...)
make build
./gwu version
make deb
make release
```

Версия задаётся только в `src/internal/gwucli/VERSION`. Аргумент `version=...`
не поддерживается. Подробнее — [сборка и Debian-пакет](docs/debian-package.md).

## Документация

- [Использование WireGuard и QoS](docs/usage.md).
- [Шаблоны конфигурации](docs/config-templates.md).
- [Архитектура кода и данные](docs/architecture.md).
- [Сборка, версии и пакетирование](docs/debian-package.md).
- [Переход с gwg и обновление gwu](docs/migration.md).
