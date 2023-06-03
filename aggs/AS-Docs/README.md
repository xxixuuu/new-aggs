# [NOTE: UNDER CONSTRUCTION]

====================================================

# BLS Aggregate (Multi-)Signature Protocol

## -Overview

    [Algorithms, Ideas:]
    (Modified) BLS (Multi-)Signature: https://crypto.stanford.edu/~dabo/pubs/papers/BLSmultisig.html

## -Description

    This is a repositry for Aggregate Signature Protocol based on Pairing Cryptosystem

Function Datagram:
![alt text](./_docs/FunctionDiagram.drawio.svg "DataFormat")

Protocol Flow:
![alt text](./_docs/AlgorithmFlow.drawio.svg "DataFormat")
## -Function Specification

### 0. Hush Funcitons

#### *(common)*

    [SHA3-256.] (powerd by golang crypto/hash):
    SHA3-256*( string s )
    A Regular SHA3-256 which is built in golang crypto package.

    [SHA3-512.] (powerd by golang crypto/hash): 
    SHA3-512*( string s )
    A Regular SHA3-512 which is built in golang crypto package.

    [Hash0.] Fouque-Tibouch Method : efp12_t* H0( string m )
    A hash fucntion could be seen as a random oracle map from a string to the addive group G2.
    .1 t = hash2base(m) : t := SHA3-512(m) mod p
    .2 w := (s * t) / (1 + B + t^2)
    .3 x1 := ((-1 + s) / 2) - t * w
    .4 x2 := -1 - x1
    .5 x3 := 1 + (1 / w^2)
    .6 e := Legendre(t)
    .7 If x1^3 + B is square, output (x1, e * sqrt(x1^3 + B))
    .8 Else if x2^3 + B is square, output (x2, e * sqrt( x2^3 + B))
    .9 Else output (x3, e * sqrt(x3^3 + B))

    [References]:
    -Section 5.3.5:
    https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-03#section-5.3
    -Sec3 Def.2:
    https://www.di.ens.fr/~fouque/pub/latincrypt12.pdf

#### *Modified BLS Aggregate Multi-Signature Protocol*

    [Hash1. (temp) Xubola Method]:
    char*[] t H1( char*[] m )
    .1 for i in range m.length t[i] := Trim(m[i], ",")
    .2 output SHA3-224(t[])

### 1. Prepareation

#### *(common)*

    [1.Gen, Q \in G2]:
    mpz_t* Sk, efp12_t* Pk GenQ( )
    1.1 Initialize BN12(BLS12) as 128-bit security.
    1.2 output BN12_Generate(Q)

    [2.Gen Sk,Pk]: 
    GenKeys( )
    2.1 Choose a random value s as Secret key: Sk
    2.2 BN12_ECSCM(s, Q) = [s]Q : Pk
    2.3 output s, [s]Q

### 2. Single Sign & Verify Protocol

#### *(common)*

    [3.Single Sign]: 
    efp12_t* T, efp12_t* σ S_Sign( char* m, mpz_t* Sk )
    3.1 T := H0(m)
    3.2 σ := ECSCMP(s, T) = [s]T = σ
    3.3 output T, σ

    [4.Single Verification]:
    bool flag S_Verify( efp12_t* Pk, char* m, efp12_t* σ )
    4.1 T := H0(m)
    4.2 If e(Q, σ) == e(Pk, T) output 1
    4.3 Else output 0

### 3. Aggregate & Verify Signature Protocol

#### *BLS Aggregate Signature Protocol*

    [5.Aggregate Signatures]:
    efp12_t* AGσ A_Sign( efp12_t*[] σs )
    5.1 For i in range σs.length AGσ := ECA(σs[])

    [All Different Message]
    [6.Aggregate Verification]:
    bool flag A_Verify(efp12_t*[] Pk, efp12_t*[] H0(m), efp12_t* AGσ )
    6.1 For i in range Pk.length check->
        if e(Q, AGσ) == e (Pk[i], H0(m)[i]) or not

    [All Same Message]
    bool flag A_Verify(efp12_t*[] Pk, efp12_t* H0(m), efp12_t* AGσ )
    6.1 For i in range Pk.length APk := ECA(Pk[])
    6.2 check -> 
        if e(Q, AGσ) == e (APk, H0(m)) or not

#### *Modified BLS Aggregate Multi-Signature Protocol*

    [5.Aggregate Signatures]:
    5.1 for i in range Pk.length T[i] := H1(Pk[i])
    5.2 tσ[i] = ECSCM(T[i], σs[i])
    5.3 σ := ECA(Tσ[])   //ECA all rational points

    [6.Aggregate Verification]:
    bool flag A_Verify(efp12_t*[] Pk, char* m, efp12_t* σ )
    6.1 for i in range Pk.length T[i] := H1(Pk[i])
    6.2 TPk[i] := ECSCM(T[i], Pk[i])
    6.3 aPk = ECA(TPk[]) //ECA all rational points
    6.4 check ->
        if e(Q, σ) == e(aPk, H0(m)) or not
    
    -Batch Aggregate & Verify Signature Verification

    [7.Batch Aggregation]:
    efp12_t* aσ B_A_Sign( efp12_t*[] σs )
    7.1 output ECA(σs[])

    [ No duplicated message m ]

    [8.Batch Verification]:
    B_Verify_nd(char*[] m, efp12_t*[] aσ, efp12_t*[] aPk)
    8.1 For i in range aPk.length check ->
        if e(Q, aσ[i]) == e(aPk[i], H0(m[i])) or not

    [ Duplicatable message m ]

    [8.Batch Verification]:
    B_Verify_d( char*[] m, efp12_t*[] aσ, efp12_t*[] aPk )
    8.1 int r[] := random() \in 2^64
    8.2 For i in range aσ.length mdσ[i] := ECSCM(r[i], aσ[i])
    8.2 For i in range aPk.length check -> 
        if e(Q, aσ[]) == e([r[i]]aPk[i], H0(m[i])) or not 

## -Structure


Datagram:
![alt text](./_docs/DataFormat.drawio.svg "DataFormat")

Protocol:
![alt text](./_docs/npro4-5.jpg "Protocol")

Structure:
![alt text](_docs/overview.png "Structure")

## -Issue
    [The rogue public-key attack]

    BLS Aggregate Signature Protocol is insecure by itself due to a rogue public-key attack,
    where an attacker registers the public key pk2:=[β]Q + (pk1)^−1∈G1, where pk1 ∈ G1 is a public key of some unsuspecting user Bob,and β ←R← Zq is chosen by the attacker.
    The attacker can then claim that both it and Bob signed some message m ∈ M by presenting the aggregate signature σ := [β]H0(m). This signature verifies as an aggregate of two signatures, one from pk1 and one from pk2, because

    e(g1,σ)=e(Q, [β]H0(m))=e([β]Q, H0(m))=e(pk1⋅pk2, H0(m)).
    
    Hence, this σ satisfies following.
    
    e(g1,σ)=e(pk1⋯pkn, H0(m1)). ---(3)

    In effect, the attacker committed Bob to the message m, without Bob ever signing m.

    Defenses. There are two standard defenses against the rogue public-key attack:

    Prove knowledge of the secret key (KOSK) [Bol03, LOS06, RY07]:
    Require every user that registers a public key to prove knowledge of the corresponding secret key. However, this is difficult to enforce in practice, and does not fit well with applications to crypto currencies [MPSW18].

    Distinct messages [BGLS03, BNN07]:
    Alternatively, require that all the messages being aggregated are distinct. This can be easily enforced by always prepending the public key to every message prior to signing. However, because now all messages are distinct, we cannot take advantage of the efficiency improvement in (3) that apply when aggregating signatures on a common message m.

## -Demo

under construction

## -Dependencies

### ---General

    [build-essensial]
    sudo apt-get install -y build-essential

    [GMP] https://gmplib.org/
    sudo apt-get install -y libgmp3-dev

    [ELiPS] https://github.com/ISecOkayamaUniv/ELiPS 
    Follow the installation instruction in ELiPS/README.md

    [Golang] 
    powerd by google

    [Redigo]
    https://github.com/gomodule/redigo

    -Server side-
    [Redis] 
    https://github.com/antirez/redis

## -Usage

    [Execute:]
    go run ./Server/SMain.go
    go run ./Client/CMain.go

## -Application

    -temp-
    
    To take the advantage of our low latency FPGA implementation and aggregate verification protocol(AVP), 
    [video or voice] casting stream would be the best practical way to use it.
    
    To achive low latency AVP, we need to modify the batch verification
    to streaming verification by following method.

    Instated of using SHA3-224,SHA3-512, we need to use (SHA3)SHAKE128, (SHA3)SHAKE256 for hash functions.

# Special thanks

## Dan Boneh, Manu Drijvers, Gregory Neven