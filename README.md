# daikin2mqtt

ダイキン製エアコンとアシストサーキュレータをMQTTで操作するためのブリッジ。

## 動作確認機種

- 無線LAN接続アダプター BRP051A41 接続のダイキン製エアコン
  (ファームウェアバージョン2.3.2)
- ダイキン製アシストサーキュレータ
  https://www.daikinaircon.com/assist-circulator/
  (ファームウェアバージョン2.3.5)

アダプタの種類やファームウェアバージョンによって、通信仕様に違いがあり
そうなので、上記と違う場合は動かないかもしれません。

## 設定ファイル

JSON形式で設定ファイルを準備します。サンプルのconfig_sample.jsonを参考。

それぞれの項目に設定する内容は以下の通り。

- type … aircon または circulator のどちらか
- name … MQTTのtopic名に使用する名前
- host … アダプタにアクセス可能なホスト名またはIPアドレス、
          もしくはサーバ経由でアクセスする場合は daikinsmartdb.jp
- id … サーバ経由の場合のID
- pw … サーバ経由の場合のパスワード
- port … サーバ経由の場合の番号

無線LAN接続アダプターのバージョンによって、直接アクセス可能なものと、
サーバ経由でアクセスするものとがあります。
(以前は全て直接アクセス可能でしたが、最近アシストサーキュレータのファー
ムウェアバージョンアップしたら、直接アクセスできなくなりました)

直接アクセス可能かどうかは

```
% curl http://エアコンのIPアドレス/common/basic_info
ret=OK,type=aircon,reg=jp,dst=0,ver=2_3_2,pow=0,err=0,location=8,name=%e3%83%aa%e3%83%93%e3%83%b3%e3%82%b0,icon=4,method=polling,port=30050,id=XXXXXXXX,pw=XXXXXXXX,lpw_flag=0,adp_kind=0,pv=2,cpv=2,led=1,en_setzone=0,mac=XXXXXXXXXXXX
```

のように返事があるかどうかで分かります。
返事がなかったら、スマートフォン用のダイキンスマートアプリで登録してから、
そのIDとパスワードを`id`と`spw`のパラメータに与えて

```
% curl 'https://daikinsmartdb.jp/common/basic_info?id=XXXXXXXX&spw=XXXXXXXX&port=30050'
ret=OK,type=aircon,ver=2_3_2,location=8,name=%E3%83%AA%E3%83%93%E3%83%B3%E3%82%B0,icon=4,method=polling,port=30050,id=XXXXXXXX,pw=XXXXXXXX,reg=jp,pow=0,err=0,adp_kind=0,pv=2,cpv=2,led=1,en_setreg=,alertm=16,on_flg=0,alerts=0
```

と返事があれば、サーバ経由での制御ができます。
複数台登録している場合は、`port`の値を30050、30051、30052…と増やして
いくと、それぞれの機器になると思います。パーセントエンコーディングされ
ている`name`で、どれが対象になっているか分かると思います。
設定ファイルの`port`はこの`port`の値です。

## 起動方法

```
% go build
```

してから

```
% ./daikin2mqtt 設定ファイル
```

## MQTT

エアコンの場合、現在の状態を以下のMQTT topicにpublishします。

- `[type]/[name]/power` 電源がONなら`on`、OFFなら`off`
- `[type]/[name]/mode` 現在の運転モード。自動は`auto`、冷房は`cool`、暖房は`heat`
- `[type]/[name]/temperature` 現在の設定温度
- `sensor/[type]/[name]/temperature` 現在の室内温度
- `sensor/[type]/[name]/humidity` 現在の室内湿度
- `sensor/[type]/[name]/outtemp` 現在の外気温

エアコンの場合、以下のMQTT topicをsubscribeし、制御を受け付けます。

- `[type]/[name]/power/set` `on`なら電源をONに、`off`ならOFFに
- `[type]/[name]/mode/set` 運転モードの変更`auto`、`cool`、`heat`のどれか
- `[type]/[name]/temperature/set` 設定温度の変更

サーキュレータの場合、現在の状態を以下のMQTT topicにpublishします。

- `[type]/[name]/power` 電源がONなら`on`、OFFなら`off`
- `[type]/[name]/fanmode` 現在の風量。`low`、`medium`、`high`のどれか

サーキュレータ場合、以下のMQTT topicをsubscribeし、制御を受け付けます。

- `[type]/[name]/power/set` `on`なら電源をONに、`off`ならOFFに
- `[type]/[name]/fanmode/set` 風量設定。`low`、`medium`、`high`のどれか

※ `[type]`と`[name]`は設定ファイルに書いたものに対応します。

※ エアコンの湿度設定は個人的にいつも「連続」でしか使ってないので、
   湿度設定には手抜きで対応してません。

※ エアコンの風量設定等もいつも同じで使っているので、対応してません。

※ サーキュレータの風量は5段階+リズムがあるのですが、うるさくて1〜3段
   階までしか使ってないので、3段階のみの対応になってます。

## HomeAssistant

HomeAssistantと連携できます。

エアコンはHomeAssistantで以下のように設定をします。

```
climate:
  - platform: mqtt
    name: "Living Aircon"
    modes:
      - "auto"
      - "cool"
      - "heat"
      - "off"
    swing_modes:
      - auto
    fan_modes:
      - auto
    current_temperature_topic: "sensor/aircon/livingroom/temperature"
    power_command_topic:       "aircon/livingroom/power/set"
    power_state_topic:         "aircon/livingroom/power"
    mode_command_topic:        "aircon/livingroom/mode/set"
    mode_state_topic:          "aircon/livingroom/mode"
    temperature_command_topic: "aircon/livingroom/temperature/set"
    temperature_state_topic:   "aircon/livingroom/temperature"
    payload_on: "on"
    payload_off: "off"
```

サーキュレータはHomeAssistantで以下のように設定をします。

```
fan:
  - platform: mqtt
    name: "Living Circulator"
    state_topic:         "circulator/livingroom/power"
    command_topic:       "circulator/livingroom/power/set"
    speed_state_topic:   "circulator/livingroom/fanmode"
    speed_command_topic: "circulator/livingroom/fanmode/set"
    payload_on:  "on"
    payload_off: "off"
    payload_low_speed:    "low"
    payload_medium_speed: "medium"
    payload_high_speed:   "high"
    speeds:
      - "low"
      - "medium"
      - "high"
```
