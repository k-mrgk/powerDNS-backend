# powerDNS-backend

研究で使用しているPowerDNSのPipeBackendです  
現在Aレコードしか対応していません 


## 機能

 - 重み付けラウンドロビン
 - 重みやTTL，IPアドレスの動的な変更

## pdns.conf

コンパイルして`/etc/powerDNS/backend/`においた場合の設定  
`pipe-regex` `cache-ttl` `query-cache-ttl`は環境に合わせて変更してください

```
launch+=pipe
pipe-command=/etc/powerdns/backend/pdns-backend
pipe-regex=^[A-Za-z]*\.sai\.test$
cache-ttl=0
query-cache-ttl=0
```


## config.json

設定ファイル  
クエリが来る度に読み込まれます  
numが2以上の場合は一番小さいTTLが使用されます
<table>
  <tbody>
    <tr>
      <td><tt>domain</tt></td>
      <td>回答したいドメイン名</td>
    </tr>
    <tr>
      <td><tt>type</tt></td>
      <td>レコードタイプ</td>
    </tr>
    <tr>
      <td><tt>num</tt></td>
      <td>回答に使うIPアドレス数</td>
    </tr>
    <tr>
    <td rowspan="3"><tt>record</tt></td>
      <td><tt>ip</tt></td>
      <td>IPアドレス</td>
    </tr>
    <tr>
    <td><tt>Weight</tt></td>
      <td>IPアドレスの重み(int)</td>
    </tr>
    <tr>
    <td><tt>TTL</tt></td>
      <td>IPアドレスのTTL(int)</td>
    </tr>
  </tbody>
</table>

