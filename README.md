# sntp

GoでSNTPクライアントの実装

NTPサーバーに対してUDPリクエストを投げて、返ってきたパケットからタイムスタンプを取得

往復の通信遅延とサーバーの処理時間を加味して、システムクロックを補正した現在の時間を返す。

```bash
$ ./sntp ntp.nict.jp

// return
2021-12-20 02:09:01.368542389 +0900 JST m=-0.465424212
```

このあたりの資料を参照
- RFC2030  
- RFC1305
- RFC768

