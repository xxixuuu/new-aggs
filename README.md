# aggs-project
<img width="401" alt="241159115-7d55070c-89e0-4ea5-917f-4f9213641275" src="https://github.com/xxixuuu/new-aggs/assets/56621296/c6ba1d5c-ba4e-42d8-b16f-fb74e024f9cb">


詳細はそれぞれのフォルダのREADME.mdを参照

以下マニュアル．　

(リポジトリとフォルダの対応関係）<br>
github.com/xxixuuu/AS-Docs　      → 　/aggs-project/aggs/AS-Docs

github.com/xxixuuu/AS-Verifier　  →   /aggs-project/AS-Proj/AS-Verifier

github.com/xxixuuu/AS-Aggregator　→   /aggs-project/AS-Proj/AS-Aggregator

github.com/xxixuuu/AS-Signer 　   →   /aggs-project/AS-Proj/AS-Signer

github.com/xxixuuu/AS-Dummy　     →   /aggs-project/AS-Proj/AS-Dummy

# TAGGS ManualTAGGS Manual

このドキュメントの対象：TAGGSSを引き継ぐ予定の方，一定の暗号知識，クラウド知識のある方．**ankoman, ANSIIRU,** ****yoshidanaokidk****

---

# Table Of Contents

# TAGGSS: Traceable AGGregate Signature Simulator

> Simulate a real IoT Network using the cloud hosting container service
> 

## 1. Concept

コンテナ技術の変わった使い方として，各コンテナ(コンテナオーケストレーターの文脈で言うとPod)を一つ仮想的なデバイスとして見立て，実際に存在するIoTネットワーク構造を模倣し，よりリアルなシミュレーションを行うことができるようにする．

**予想されるメリット**：現在稼働中の本番環境をいじることなく，低コストで既存ネットワーク全体に対して大規模な新規技術の導入＆チューニング，最適化探索シミュレーションができるようになる．

**予想されるデメリット**：仮想化レイヤは薄くても，コンテナオーケストレーションレイヤの分厚さがわからないため，実システムの挙動と比べた時の差分が不透明．同じネットワークをモデル化してシミュレーションを行った時(ns3など)との差分，効果，コスパなど不透明．コンテナのデプロイに時間がかかるので，大規模なネットワークを模倣する場合では時間がとてもかかる場合がある．

・レイヤ分類

| Theory | Papers |
| --- | --- |
| Application | golangプログラム |
| Simulation | AWS, K8S |

**キーワード**：コンテナ，golang, TCP通信, Traceable Aggregate Signature, Simulator, NoSQL, AWS, EKS, k8s, Container, Kind, IoT Network

## 2. Assets

注意：コードがわからないときは聞いた方が早いことが多い．特にAggregatorの中は並行性を保つためにスパゲッティになっているため，コードを追いかけると迷うことが多い．ただ，TAGGSSのなかでAGGの三文字を占めていることからわかるように，Aggregatorがこの話のキモである．

引き継ぎで始めに見ていくリポジトリの順番もこの↓順番

- **github.com/xxixuuu/AS-Docs**

図，エンティティ間の通信フォーマット，インフラ構成，eksctl, k8s, ECR制御用のスクリプトなどが(以下のすべてのリポで)”書き散らかされ”ている．(すまん)

以下構造：

- **Application Source Code**
    - github.com/xxixuuu/AS-Verifier
    - github.com/xxixuuu/AS-Aggregator
    - github.com/xxixuuu/AS-Signer
    - github.com/xxixuuu/AS-Dummy

フルスクラッチのAGGSエンティティソースコード(Dummy, Signer, Aggregator, Verifier)，複数のBranchがあるが，commitが一番新しいものは基本的に動く, build folderのなかのものが最新

- **github.com/xxixuuu/AS-Dump**

TAGGSでシミュレーション実行後に結果としてRedis(DB)に保存されるデータを抽出してcsvなどにするプログラム，工夫すれば平均取ったり，分布取ったりできる．プログラミングの腕次第．

- **go library(middleware)**
    - github.com/xxixuuu/netcp
    - github.com/xxixuuu/utils
    - ~~github.com/xxixuuu/kvs~~

Application Source Code中で呼び出されるお手製依存ライブラリ，go制．主にgoのstandard libraryを使ってこちらで設計したTCP通信(netcp)，プログラミング中の型とTCPに書き込むためのBytes型のフォーマット変換(utils)，RedisのREST APIが定義(kvs，こっちはverifierとmergeしたかも)されている．

- **github.com/xxixuuu/AS-Test**

アプリケーションのテストコードではなく，仕組みなどを実装するときに試しで組む実験コードが書き散らかされている．

**外部依存ライブラリ**：herumi/bls-go-binary, libssl-dev ...もっとあるかもだが，重要なのはこの二つ

Terminalのおすすめ：Windows Terminal (Window Splitできる)

画面のおすすめ：ソースコードと動作プログラムを複数監視する必要があるので，49インチのphillipsのあれ(+24インチx2)はあった方がいい，空いてなければもう一台買っといたほうがいい．24インチx3でもできるが，少し物足りない．

**misc** : ****自分の過去スライド，分かりにくいけど参考にしてくれぇ・・・

## 3. メンテナンス

レイヤごとのメンテナンス

## 4. シミュレーションの実行

### ・基本的なAGGSプログラムの動作チェック

**前提： Redis, Dockerのローカルインストール, すべてのソースコードのダウンロード**

0-1 : goのインストール [https://go.dev/doc/install](https://go.dev/doc/install)

0-2 : ソースコードのダウンロード(AS-Dummy, AS-Signer, AS-Aggregator, AS-Verifier/v2)

```bash
git@github.com:xxixuuu/AS-Verifier.git
git@github.com:xxixuuu/AS-Aggregator.git
git@github.com:xxixuuu/AS-Signer.git
git@github.com:xxixuuu/AS-Dummy.git
```

**各リポジトリの中でも，/buildフォルダ以降のソースコードがSCIS2022のデータ収集に使用した最新の実装である．**

0-3 : Redis のインストール [https://redis.io/download/](https://redis.io/download/)

Redis の起動 

```bash
redis-server
```

Redis-cli の起動 (新しい Termial Window)

```bash
redis-cli
```

すると、サーバ情報が "[IPaddress]:[Port]>" という書式で表示されるのでその値を覚えておく

```
127.0.0.1:6379> keys *
```

KVS = Key Value Store

0-4, 0-7 の起動時の引数はAS-Docs/Scripts/AWS/2-k8s Level/pods/*.yamlのargs: ["10.100.84.121","3000","3000","0","500","64","1"]みたいなのが丸ごとつかえる

起動時の引数の仕様は AS-Docs\Scripts\AWS\0-SW Level\BasicStartupCommands [copy.sh](http://copy.sh/)にある．

0-4 : Verifierの起動 (新しい Termial Window)

```bash
sudo apt-get install libssl-dev # 事前に Openssl のインストールを済ませる(ubuntの場合）
cd AS-Verifier/build
go run VMain.go localhost 3000 0 #port番号は6397以外
go run IPaddress Port [,mode(v2の場合)] #bash run.sh 127.0.0.1 6397 0

#Verifier in following arguments  "-" = use setting on json file
#[1] = RedisAddress
#[2] = ListenPort
#new [3] = Tracking Method

#-VIR started in mode: 0
#[2022-04-12 14:08:09.3080546 +0900 JST m=+0.001040401]
#-VIR: Using maxmum cpu core up to :12
#PONG <nil>
#-VIR: tcp4 handler listning on 0.0.0.0:6379

#[I'm all ears]
```

0-5 : Aggregatorの起動 (新しい Termial Window)

```bash
cd AS-Aggregator/build
go run AMain.go localhost 3000 3001 0 500 64 1　
#[1] = ParentAddress
#[2] = ParentPort 
#[3] = ListenPort
#[4] = Mode 0 -Timewindow aggregation, 1 -Count Threashold aggregation,  2 -Mixed Aggregation
#new [5] = Time window
#new [6] = Signer Handler
#new [7] = Tracking Method
```

0-6 : Signerの起動 (新しい Termial Window)

```bash
cd AS-Signer/build
go run SMain.go localhost 3001 3002 0 - 1
#Signer in following arguments "-" = use setting on json file
#[1] = AggregatorAddress
#[2] = AggregatorPort
#[3] = ListenPort
#[4] = Mode:     0 normal mode, 1 false mode, 2 manual timing false mode,  (only work on obsolete multi-thread mode) 3 normal -> false mode with falltime,
#[5] = TimeTillFall
#[6] = SID
```

0-7 : Dummyの起動 (新しい Termial Window)

```bash
cd AS-Dummy/build
go run DMain.go - 3002 0 10 5 10000 #Unix Time

#Dummy in following arguments "-" = use setting on json file sync
#[1] = SignerAddress
#[2] = SignerPort
#[3] = Mode:     0 = timeStamp, 1 = designated length randomBytes, 2 = designated length randomBytes with range    
#[4] = MessageInterval
#[5] = DataLength
#new [6] = MaxDataOutPut
```

0-8 : データの確認 (新しい Termial Window)

```bash
redis-cli
keys *
```

### **ローカルでのAGGS Application動作チェック(Kind Cluster)：**

**Docker imageの準備**

・redis image の pull 

```bash
docker pull redis
```

・Dummyのdocker imageのビルド

```bash
cd AS-Dummy
cd build
bash build.sh 1
```

・(image buildの確認)

```bash
docker images
```

・同様に Signer, Aggregator, Verifierのイメージもビルドする(bash build.sh 1) *v2 ブランチ

・Kind Clusterの準備

```bash
cd AS-Docs/Scripts/KIND/small
bash clusterTest.sh

#kind get nodes でクラスタが存在した場合は
#kind delete cluster　で削除
#kubu-universeのインストール(https://github.com/afritzler/kube-universe)
```

・(Clusterの起動確認)

```bash
(sudo) kubectl get nodes
```

・Serviceの配置

```bash
cd ../service
(sudo) kubectl apply -f .
```

・Serviceの配置の確認及びServiceの内部ipを確認

```bash
(sudo) kubectl get svc
#(example)
#NAME                  TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
#kubernetes            ClusterIP   10.96.0.1       <none>        443/TCP    8m6s
#signer-aggregator     ClusterIP   10.98.117.29    <none>        3000/TCP   21s
#verifier-aggregator   ClusterIP   10.102.154.47   <none>        3000/TCP   21s
#verifier-database     ClusterIP   10.110.31.238   <none>        6379/TCP   21s
```

(各serviceのipをここでメモっておく)

・例) signer-aggregator : 10.0.0.101

・例) verifier-aggregator : 10.0.0.11

・例) verifier-database: 10.0.0.2

・Redis PodのDeploy

```bash
cd ../Pods/redis
(sudo) kubectl apply -k.
```

・Redis PodのDeployの確認 

```bash
(sudo) kubectl get pods
```

・各buildしたAGGSエンティティのPodのDeploy

・VerifierのPodのDeploy

```bash
cd ../verifier
(change verifier distination ip by editing verifier.yaml L16 ipaddress to the memoed svc verifier-database ip)
(sodo) kubectl apply -f .
```

・Aggregatorも同様

・Signer-DummyのPodのDeploy

```bash
cd ../../deployment
(change sginer distination ip by editing DepsignerDummy.yaml L26 ipaddress to the memoed svc signer-aggregator ip)
(sudo) kubectl apply -f .
(change signer-dummy number by editing replicas)
```

・動作確認　(logの確認）

```bash
(sudo) kubectl get pods -o wide     #check status of pods
         
(sudo) kubectl logs verifier-aggs   #check logs of verifier
#TRUE- msg num: 196
#getSPKVecElapsed: 18.942µs
#TRUE- msg num: 197
#VUID:198  Elapsed:82.484685ms

(sudo) kubectl logs aggregator-aggs #check logs of aggregator
(sudo) kubectl logs signer-aggs-* 
```

### **AWS上でのTAGGSS動作：AWS K8S Cluster**

**準備1： 環境．**

**1-1 : AWSアカウントセットアップ(AWS CLI Consoleのインストール)**

**1-2 : (Option)本番TAGGSS用IAMの作成**

・AWS CLI Console と　IAMアカウントの作成

下記を参照してIAMアカウントと本番TAGGSS用IAMの作成．

*”Udemyメディア，AWS CLIとは？インストール手順や使い方を初心者向けに解説[https://udemy.benesse.co.jp/development/system/aws-cli.html](https://udemy.benesse.co.jp/development/system/aws-cli.html)（最終閲覧2022/06/01)”*

・AWS CLIの設定(アクセスキーIDなどはIAMアカウント作成時のCSVファイルを参照)

```bash
aws configure    
#AWS Access Key ID [None]: (アクセスキーID)
#AWS Secret Access Key [None]: (秘密鍵)
#Default region name [None]: (ap-northeast-1など)
#Default output format [None]: json
```

・これで，WSL2のUbuntuなどからコマンドでAWS上のリソースを操作可能．

**1-3 : 本番用ECR(Elastic Container Registry)のセットアップ(リージョンはap-northeast-1)**

・localにAGGSのプログラムダウンロード（KINDと一緒）

```bash
mkdir ~/AS-Proj   #make AS-Proj in home directory 
cd ~/AS-Proj 
(clone  AS-Dummy AS-Signer AS-Aggregator AS-Verify in AS-Proj and checkout v2_branch) 
mkdir ~/aggs
(clone  AS-Docs in aggs) 
```

・AWS上にContainer image をpush ．

```bash
aws ecr create-repository --repository-name aggs2 --region ap-northeast-1 #create repository
(copy repositoryUri)
aws ecr get-login-password | docker login --username AWS --password-stdin repositoryUri(057607951461.dkr.ecr.ap-northeast-1.amazonaws.com/aggs2

cd ~/aggs/AS-Docs/Scripts/AWS/1-Docker Level
bash buildAll.sh #build dockerimage of Verify,Aggregator,Signer and Dummys 
(change repositoryUri editing by tagAll.sh and pushAll.sh L1, L2, L3 and L4 to  repositoryUri copied )
bash tagAll.sh 
bash pushAll.sh
(if you can't push dockerimages to ECR repository, run uptoECR*.sh one by one.)
```

**1-3 : Eksctlのセットアップ(ローカルにインストールするだけ)**

・下記を参照してEksctlをinstall

*”AWSドキュメントAmazon Eks ユーザガイド，Eksctlのインストール*

[*https://docs.aws.amazon.com/ja_jp/eks/latest/userguide/eksctl.html](https://docs.aws.amazon.com/ja_jp/eks/latest/userguide/eksctl.html)（最終閲覧2022/06/01)”*

**1-4 : Container imageのECSへのupload**

 1-3で完了？

**1-5 : EKS Clusterの作成**

・AWS上にEKS Clusterの作成

下記のURL参照してEKS Cluster

*“AWSドキュメントAmazon Eks ユーザガイド，Amazon EKS クラスターの作成，*

[*https://docs.aws.amazon.com/ja_jp/eks/latest/userguide/create-cluster.html](https://docs.aws.amazon.com/ja_jp/eks/latest/userguide/create-cluster.html)(最終閲覧2022/06/01)”*

```bash
eksctl create cluster --name ekd_cluster_name --version *.** --without-nodegroup
```

・cluster の作成 

```bash
eksctl create cluster --name cluser_name --version *.** --without-nodegroup
cd ~/aggs/AS-Docs/Scripts/AWS/3-AWS Level/cluster1
(chage cluster name editing by cluster.yml L5 to your eks_cluster_name)
eksctl create nodegroup --config-file=./cluster.yaml　#create nodegroup
```

#ここからはほぼKINDでの動作と同じ

・Serviceの配置

```bash
cd ../../2-K8s Level/service
kubectl apply -k.
```

・Serviceの配置の確認及びServiceの内部ipを確認

```bash
kubectl get svc
#(example)
#NAME                  TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
#kubernetes            ClusterIP   10.96.0.1       <none>        443/TCP    8m6s
#signer-aggregator     ClusterIP   10.98.117.29    <none>        3000/TCP   21s
#verifier-aggregator   ClusterIP   10.102.154.47   <none>        3000/TCP   21s
#verifier-database     ClusterIP   10.110.31.238   <none>        6379/TCP   21s
```

(各serviceのipをここでメモっておく)

・Redis PodのDeployの確認 

```bash
cd ../pods/redis
(sudo) kubectl apply -k.
```

・各buildしたAGGSエンティティのPodのDeploy

・VerifierのPodのDeploy

```bash
cd ../verifier
(change verifier distination ip by editing verifier.yaml L16 ipaddress to the memoed svc verifier-database ip)
(change image name editing by verifier.yaml L10 repositoryUri to verifier repositoryUri)
(sodo) kubectl apply -f .
```

・Aggregatorも同様

・Signer-DummyのPodのDeploy

```bash
cd ../ds-template
(change sginer distination ip by editing genPodsCL1.go to the memoed svc signer-aggregator ip)
(change image name editing by editing genPodsCL1.go  repositoryUri to signer repositoryUri)
(statement4 の argsの最後に10000(#new [6] = MaxDataOutPut)などの数字を追加する必要がある？）
go run genPodsCL1.go 4 (The number of Signer)
cd 4-CL1
(sodo) kubectl apply -f .
```

・動作確認　(logの確認）

```bash
(sudo) kubectl get pods -o wide     #check status of pods
(sudo) kubectl logs verifier-aggs   #check logs of verifier
#TRUE- msg num: 196
#getSPKVecElapsed: 18.942µs
#TRUE- msg num: 197
#VUID:198  Elapsed:82.484685ms(こういう感じのLogが出てたら動いてはいる）
(sudo) kubectl logs aggregator-aggs #check logs of aggregator
(sudo) kubectl logs signer-aggs-* -c signer #check logs of signer
```

・EKSの削除（EKSの起動時間で課金されるのでこれを必ずやる）
```bash
kubectl get svc --all-namespaces
kubectl delete svc <service-name> #各サービスを削除
kubectl -n my-ns delete pod,svc --all 
kubectl delete pod --all, svc --all
eksctl delete cluster --name <prod>
```

