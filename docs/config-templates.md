# Шаблоны конфигураций

Текстовые шаблоны встроены в бинарник через go:embed. При первом использовании
они создаются на диске; существующие файлы не заменяются обновлением пакета.

## WireGuard

Каталог: `/etc/wireguard/gwu/wireguard/templates/`.
Имена: `server.conf.tmpl`, `client.conf.tmpl`, `gateway.conf.tmpl`.

```bash
sudo gwu wireguard template show server
sudo gwu wireguard template set server --file ./server.conf.tmpl
sudo gwu wireguard template reset server
```

Вместо server доступны client и gateway. Reset явно заменяет выбранный локальный
шаблон встроенным; set заменяет содержимым файла. Эти команды сами не
регенерируют рабочий конфиг, не применяют sysctl и не перезапускают WireGuard.
Show может первоначально создать отсутствующие шаблоны.

Set/reset проверяют синтаксис Go template, пробный рендеринг и синтаксис shell
hooks через `bash -n`, без исполнения. Это не проверка сетевой семантики или
произвольных условных веток. Локальный шаблон — доверенный код администратора:
hooks будут исполняться при реальном запуске wg-quick.

### Данные шаблонов

| Шаблон | Поля |
| --- | --- |
| server | ServerPrivateKey, ServerPublicKey, LocalAddress, PublicAddress, ListenPort, Eth, Alias, DnsResolv, Users |
| client и элементы Users | ClientPrivateKey, ClientPublicKey, ClientLocalAddress, ServerPublicKey, ServerIp, ServerPort, DnsResolv, Name, Status |
| gateway | Получает серверную модель; встроенный шаблон задаёт net.ipv4.ip_forward=1 |

Status равен `active` для активного клиента и пустой строке для заблокированного.
Серверный шаблон использует `{{ range .Users }}`.
Функция `natTable .Alias` строит безопасное имя nftables-таблицы.

Исходные шаблоны:

- [server.conf.tmpl](../src/modules/wireguard/templates/server.conf.tmpl).
- [client.conf.tmpl](../src/modules/wireguard/templates/client.conf.tmpl).
- [gateway.conf.tmpl](../src/modules/wireguard/templates/gateway.conf.tmpl).

Встроенный server template создаёт только NAT в отдельной nftables-таблице.
FORWARD-разрешения и source rules в нём отсутствуют. Переход со старых
iptables hooks требует отдельного согласования с установленным firewall.

При следующей операции пользователя серверный конфиг будет сгенерирован из
текущего шаблона. Изменение client template влияет на новых клиентов и не
переписывает уже выданные конфиги. Gateway применяется при системной подготовке
init; не запускайте init для обновления существующего сервера.

## QoS

Каталог: `/etc/wireguard/gwu/qos/templates/`.
Файлы: `channel.json.tmpl`, `qos.service.tmpl`.

`gwu qos init --interface wg0 --rate 75Mbit` создаёт отсутствующий channel.json,
проверяет его и генерирует unit. `gwu qos render` перечитывает реальный JSON
и локальный шаблон unit, не меняя JSON и работающие очереди.
Повторный init сохраняет существующий channel.json.

Шаблонам доступны Interface, Rate и ConfigPath.
Unit записывается в `/etc/wireguard/gwu/qos/generated/qos.service`.
Генерация не означает установку или включение unit.

Metadata WireGuard, рабочая QoS-конфигурация и ownership state хранятся как
типизированный JSON; текстовые шаблоны не заменяют их модели.
