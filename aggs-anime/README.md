# 集約署名シミュレーションの可視化

## 必要ソフトウェア

- Node.js
  - version 16 以上
  - https://nodejs.org/

## プログラムのビルド

```bash
npm install
npm run build
```

## プログラムの動作確認

ビルドしたプログラムは `dist` 以下に保村されます。
ブラウザで動作するアプリケーションのため、Python 等で Web サーバーを起動すると確認しやすいです。

```bash
cd dist
python -m http.server
# ブラウザで http://localhost:8000 を開いて確認
```

## プログラムの編集

### サンプルデータの変更

サンプルデータは `src/samples` 以下に置いてあります。
JSON を編集後は再度ビルドしてください。
