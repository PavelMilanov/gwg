# Использование gwu

## Первоначальная настройка WireGuard

Административные команды выполняйте через sudo. Справка, version и генерация
completion доступны без root.

Перед настройкой проверьте default route, доступность endpoint сервера и
firewall. Разрешите входящий UDP-порт WireGuard, forwarded-трафик VPN-клиентов
и обратные ответы в существующем firewall. Gwu не создаёт filter rules.
При нескольких default routes или NAT перед сервером проверьте автоматически
определяемые внешний интерфейс и endpoint в metadata и конфиге.

На новом сервере без существующего wg0:

```bash
sudo gwu wireguard init
sudo wg show wg0
sysctl net.ipv4.ip_forward
sudo nft list table ip gwu_nat_776730
```

Команда создаёт wg0 с адресом 10.0.0.1/24 и UDP-портом 51830, включает
IPv4 forwarding, генерирует ключи и включает/запускает wg-quick@wg0.
DNS в новых клиентских конфигурациях по умолчанию — 8.8.8.8.
Локальный DNS-сервер не устанавливается.

Если система уже подготовлена и нужен другой интерфейс, вместо init:

```bash
sudo gwu wireguard server install --name vpn7 --network 10.7.0.1/24 --port 51900
```

Server install не выполняет системную подготовку init: IPv4 forwarding и
firewall должны быть настроены заранее. Не используйте обе команды подряд
для создания двух управляемых серверов: пользовательские операции ожидают
единственный JSON сервера.

При ошибке настройки проверьте состояние unit, `wg show` и nftables.
Init может успеть записать системные настройки до отказа установки сервера.
Не повторяйте его как способ исправления действующего интерфейса.
Полного автоматического отката сетевой настройки нет.

## Пользователи

```bash
sudo gwu wireguard user add alice
sudo gwu wireguard user list
sudo gwu wireguard user block alice
sudo gwu wireguard user unblock alice
sudo gwu wireguard user remove alice
```

Это отдельные примеры операций, а не последовательность настройки.
Add/remove/block/unblock регенерируют серверный конфиг и **перезапускают
WireGuard**, что может прервать соединения. List не меняет peers.

Конфиг нового клиента:
`/etc/wireguard/gwu/wireguard/configs/alice.conf`.
Metadata: `/etc/wireguard/gwu/wireguard/clients/alice.json`.
Оба файла содержат секреты. Не публикуйте их и не включайте в отчёты диагностики.

Для диагностики:

```bash
sudo wg show wg0
systemctl status wg-quick@wg0.service
```

Команд server show/stat в gwu нет.

## Общий QoS

QoS — прототип общего ограничения исходящего трафика интерфейса,
то есть download VPN-клиентов. Индивидуальных полос, ACL и ограничения upload нет.

На уже работающем wg0 с отсутствующей root qdisc либо каналом, ранее созданным gwu:

```bash
sudo gwu qos init --interface wg0 --rate 75Mbit
sudo gwu qos render
sudo gwu qos apply --config /etc/wireguard/gwu/qos/channel.json
sudo gwu qos show --interface wg0
sudo tc -s qdisc show dev wg0
sudo tc -s class show dev wg0
```

Init/render только генерируют файлы; apply меняет TC через Netlink.
Если channel.json уже существует, повторный init сохраняет его скорость.
Для изменения редактируйте `/etc/wireguard/gwu/qos/channel.json` через vim:

```json
{
  "interface": "wg0",
  "rate": "75Mbit"
}
```

Повторный apply обновляет принадлежащий gwu класс. Чужие очереди, в том числе
созданные старым tc.service, не захватываются и не удаляются.

### Необязательный systemd unit

После qos init просмотрите целиком сгенерированный unit:

```bash
sudo vim /etc/wireguard/gwu/qos/generated/qos.service
```

Убедитесь, что пути и интерфейс соответствуют конфигурации и что
`/etc/systemd/system/qos.service` ещё не занят другой службой.
Затем для включения автоматического применения:

```bash
sudo install -m 0644 /etc/wireguard/gwu/qos/generated/qos.service /etc/systemd/system/qos.service
sudo systemctl daemon-reload
sudo systemctl enable --now qos.service
sudo gwu qos show --interface wg0
```

Полный исходный unit: [qos.service.tmpl](../src/modules/qos/templates/qos.service.tmpl).
Он связан с wg-quick unit. После изменения channel.json:

```bash
sudo systemctl reload qos.service
sudo gwu qos show --interface wg0
```

Reload/restart qos не перезапускает WireGuard. У unit нет ExecStop, поэтому
остановка не удаляет очередь. Чтобы отменить включение этого unit и убрать
принадлежащее gwu ограничение:

```bash
sudo systemctl disable --now qos.service
sudo gwu qos down --interface wg0
sudo tc -s qdisc show dev wg0
```

Это снимает shaping, но не восстанавливает прежний сторонний QoS.
Down откажет, если не может подтвердить владение. При ошибке сохраните state
и исследуйте расхождение с Netlink; не удаляйте state или чужую root qdisc
для обхода проверки.

## Bash completion

Debian-пакет устанавливает
`/usr/share/bash-completion/completions/gwu`.
При установленном bash-completion откройте новую Bash-сессию.

Для просмотра генерируемого скрипта без изменения shell:

```bash
gwu completion bash
```

Команда completion скрыта из основной справки. Пакет не меняет .bashrc и aliases.
