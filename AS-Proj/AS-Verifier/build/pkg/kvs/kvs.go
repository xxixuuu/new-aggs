package kvs

import (
	"context"
	"encoding/binary"
	"encoding/gob"
	"fmt"

	redis "github.com/go-redis/redis/v8"
	"github.com/xxixuuu/utils"
)

// Signer --
type Signer struct {
	KEY string //key sr:Sprint(AID):Sprint(SID)

	IP      string
	PORT    string
	AID     []byte
	SID     []byte
	SPK     []byte //Signer Public Key
	SPKSIGN []byte //Self Sign
	REGDATE []byte //mostly 8byte
}

// Aggregator --
type Aggregator struct {
	KEY     string //key ag:Sprint(AID)
	IP      string
	PORT    string
	AID     []byte
	SID     []uint64 //append
	REGDATE []byte
}

// MessageData --
type MessageData struct {
	KEY string //md:int64(AID):int64(SID):int64(MID)

	STATEHEADER        []byte //mostly means statas
	AID                []byte
	SID                []byte
	MESSAGEHEAHDER     []byte
	MESSAGE            []byte //mostly means Gen TS
	HashedMessage      []byte //by sha512
	SIGNEROPTIONHEADER []byte
	SIGNEROPTION       []byte //mostly means signed TS
	MID                uint64
	FWDDATE            []byte //Only needed when debug mode(aggregator)
	REGDATE            []byte //mostly means Arrival at Verifier TS //Arrival
	SIGN               []byte
}

func (md *MessageData) TrimSignatureScope(SIDAssign int) []byte {
	if SIDAssign == 1 {
		return append(append(append(md.MESSAGEHEAHDER, md.MESSAGE...), md.SIGNEROPTIONHEADER...), md.SIGNEROPTION...)
	}
	return append(append(append(append(md.SID, md.MESSAGEHEAHDER...), md.MESSAGE...), md.SIGNEROPTIONHEADER...), md.SIGNEROPTION...)
}

type HashedMessageWithID struct {
	SID           []byte
	HashedMessage []byte
}

type HashedMessageArraywithID struct {
	VUID               int
	HashedMessageArray []*HashedMessageWithID
}

func (ts *HashedMessageArraywithID) Print() {
	println("VUID", ts.VUID)
	for i := range ts.HashedMessageArray {
		print("SID:", ts.HashedMessageArray[i].SID, ",")
	}
}

func (ts *HashedMessageArraywithID) AppendNew(SID []byte, HashedMessage []byte) {
	newHashedMessageWithID := new(HashedMessageWithID)
	newHashedMessageWithID.SID = SID
	newHashedMessageWithID.HashedMessage = HashedMessage
	ts.HashedMessageArray = append(ts.HashedMessageArray, newHashedMessageWithID)
}

func InitRoundBufferArray(hma *[]*HashedMessageArraywithID) {
	for i := len(*hma) - 1; i > 0; i-- {
		(*hma)[i] = new(HashedMessageArraywithID)
	}
}

func UpdateRoundBufferArray(hma *[]*HashedMessageArraywithID) {
	for i := len(*hma) - 1; i > 0; i-- {
		(*hma)[i] = (*hma)[i-1]
	}
}

type SortBy []*HashedMessageArraywithID

func (a *HashedMessageArraywithID) Len() int { return len(a.HashedMessageArray) }
func (a *HashedMessageArraywithID) Swap(i, j int) {
	a.HashedMessageArray[i], a.HashedMessageArray[j] = a.HashedMessageArray[j], a.HashedMessageArray[i]
}
func (a *HashedMessageArraywithID) Less(i, j int) bool {
	return utils.ByteArrayConvertToUint(a.HashedMessageArray[i].SID) < utils.ByteArrayConvertToUint(a.HashedMessageArray[j].SID)
}

func (ts *HashedMessageArraywithID) TrimZeroToIndex(index int) {
	newArray := make([]*HashedMessageWithID, 0)
	for i := index; i < len(ts.HashedMessageArray); i++ {
		newArray = append(newArray, ts.HashedMessageArray[i])
	}
	ts.HashedMessageArray = newArray
}

func (ts *HashedMessageArraywithID) TrimSingleSignerMessageArrayScope() (res [][]byte) {
	// for i := range ts.HashedMessageArray {
	// 	print("SID:", utils.ByteArrayConvertToUint(ts.HashedMessageArray[i].SID), ",")
	// }
	index := 0
	for i := 0; i < len(ts.HashedMessageArray)-1; i++ {
		if utils.ByteArrayConvertToUint(ts.HashedMessageArray[i].SID) != utils.ByteArrayConvertToUint(ts.HashedMessageArray[i+1].SID) {
			index = i
			break
		}
		if i == len(ts.HashedMessageArray)-2 {
			index = i + 1
		}
	}
	index++
	res = make([][]byte, 0)
	for i := 0; i < index; i++ {
		res = append(res, ts.HashedMessageArray[i].HashedMessage)
	}
	ts.TrimZeroToIndex(index)
	return res
}

// VerifyData --
type VerifyData struct {
	KEY     string //vd:[AID]:[VDID]
	AID     []byte
	VDID    uint64
	MID     []uint64
	HASHVEC [][]byte //hash1, hash2
	AGS     []byte   //Aggregated Signature
	RES     bool     //result
	AGGDATE []byte   //AGS Generated TS by Aggregator stamp
	REGDATE []byte   //Verify TS
}

type Feedback struct {
	AID         []byte
	SID         []byte
	VUID        []byte
	REGDATE     []byte
	VSIGN       []byte
	TRACKEDDATE []byte
}

// NewClient --
func NewClient(ctx context.Context, dbip string) *redis.Client {
	kvs := redis.NewClient(&redis.Options{
		Addr:     dbip + ":6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := kvs.Ping(ctx).Result()
	fmt.Println(pong, err)
	// Output: PONG <nil>
	return kvs
}

// SetSigner --
func (signer *Signer) Set(ctx context.Context, db *redis.Client) {
	err := db.HSet(
		ctx,
		"sr"+":"+fmt.Sprint(utils.ByteArrayConvertToUint(signer.AID))+":"+fmt.Sprint(utils.ByteArrayConvertToUint(signer.SID)),
		"ip", signer.IP,
		"pt", signer.PORT,
		"sid", string(signer.SID),
		"aid", string(signer.AID),
		"spk", string(signer.SPK),
		"slfsig", string(signer.SPKSIGN),
		"rd", string(signer.REGDATE),
	).Err()
	if err != nil {
		println("KVS HSet() Error:")
		fmt.Println(err)
	}
}

// GetSigner --
func (signer *Signer) Get(ctx context.Context, db *redis.Client) {

	val, err := db.HGetAll(ctx, "sr"+":"+fmt.Sprint(signer.AID)+":"+fmt.Sprint(signer.AID)).Result()
	if err != nil {
		panic(err)
	}
	signer.IP = val["ip"]
	signer.PORT = val["pt"]
	signer.SID = []byte(val["sid"])
	signer.AID = []byte(val["aid"])
	signer.SPK = []byte(val["spk"])
	signer.SPKSIGN = []byte(val["slfsgn"])
	signer.REGDATE = []byte(val["rd"])

}

func (signer *Signer) SetFromMessageData(md *MessageData, SID []byte, AID []byte, assignFlag int) {

	signer.IP = ""   //think about it later
	signer.PORT = "" //think about it later
	signer.AID = AID
	// if assignFlag == 1 {
	// signer.SID = SID //or SID
	// } else if assignFlag == 0 {
	signer.SID = md.SID //or SID
	// }
	signer.SPK = md.MESSAGE
	signer.SPKSIGN = md.SIGN
	signer.REGDATE = utils.NowUnixNanoLittleEndian()

}

// SetAggregator --
func (aggregator *Aggregator) Set(ctx context.Context, db *redis.Client) {
	aid := binary.LittleEndian.Uint32(aggregator.AID)
	err := db.HSet(
		ctx,
		"ag"+":"+fmt.Sprint(aid),
		"ip", aggregator.IP,
		"pt", aggregator.PORT,
		"aid", string(aggregator.AID),
		"rd", string(aggregator.REGDATE),
		"sid", utils.GobEncoderOutString(aggregator.SID),
	).Err()
	if err != nil {
		println("KVS HSet() Error:")
		fmt.Println(err)
	}
}

// GetAggregator --
func (aggregator *Aggregator) Get(ctx context.Context, db *redis.Client) {

	val, err := db.HGetAll(ctx, "ag"+":"+fmt.Sprint(utils.ByteArrayConvertToUint(aggregator.AID))).Result()
	if err != nil {
		println("KVS HGet() Error:")
		fmt.Println(err)
	}
	aggregator.IP = val["ip"]
	aggregator.PORT = val["pt"]
	aggregator.AID = []byte(val["aid"])
	aggregator.SID = utils.GobDecoderUint64Array(val["sid"])
	aggregator.REGDATE = []byte(val["rd"])

}

// AppendSigner --
func (aggregator *Aggregator) Append(ctx context.Context, db *redis.Client, signer *Signer) {

	aggregator.Get(ctx, db)
	aggregator.SID = append(aggregator.SID, uint64(utils.ByteArrayConvertToUint(signer.SID)))
	go aggregator.Set(ctx, db) //update

	fmt.Println("New Signer", signer.SID, " Append to Aggregator AID", aggregator.AID)
}

func (aggregator *Aggregator) SetFromMessageData(md *MessageData, AID []byte, assignFlag int, ip string, port string) {
	if assignFlag == 0 {
		aggregator.AID = md.AID
	} else if assignFlag == 1 {
		aggregator.AID = AID
	}

	aggregator.IP = ip
	aggregator.PORT = port
	aggregator.REGDATE = utils.NowUnixNanoLittleEndian()
}

// SetMessage --
func (message *MessageData) Set(ctx context.Context, db *redis.Client) {

	err := db.HSet(
		ctx,
		"md:"+fmt.Sprint(utils.ByteArrayConvertToUint(message.AID))+":"+fmt.Sprint(utils.ByteArrayConvertToUint(message.SID))+":"+fmt.Sprint(message.MID),
		"aid", string(message.AID),
		"sid", string(message.SID),
		"mh", string(message.MESSAGEHEAHDER),
		"m", string(message.MESSAGE),
		"soh", string(message.SIGNEROPTIONHEADER),
		"so", string(message.SIGNEROPTION),
		"mid", fmt.Sprint(message.MID),
		"fd", string(message.FWDDATE),
		"rd", string(message.REGDATE),
	).Err()
	if err != nil {
		println("KVS HSet() Error:")
		fmt.Println(err)
	}
}

// GetMessage --
func (message *MessageData) Get(ctx context.Context, db *redis.Client) {

	val, err := db.HGetAll(ctx, "md:"+fmt.Sprint(message.AID)+":"+fmt.Sprint(message.SID)+":"+fmt.Sprint(message.MID)).Result()
	if err != nil {
		println("KVS HGet() Error:")
		fmt.Println(err)
	}
	// aggregator.IP = val["ip/"]
	message.AID = []byte(val["aid"])
	message.SID = []byte(val["sid"])
	message.MESSAGEHEAHDER = []byte(val["mh"])
	message.MESSAGE = []byte(val["m"])
	message.SIGNEROPTIONHEADER = []byte(val["soh"])
	message.SIGNEROPTION = []byte(val["so"])
	message.MID = uint64(utils.GetUint64FromString2Map("mid", val))
	message.FWDDATE = []byte(val["fd"])
	message.REGDATE = []byte(val["rd"])
}

// SetVerifyData --
func (verifiedData *VerifyData) Set(ctx context.Context, db *redis.Client) {
	gob.Register([][]uint8(verifiedData.HASHVEC))

	err := db.HSet(
		ctx,
		"vd:"+fmt.Sprint(utils.ByteArrayConvertToUint(verifiedData.AID))+":"+fmt.Sprint(verifiedData.VDID),
		"aid", string(verifiedData.AID),
		"vdid", fmt.Sprint(verifiedData.VDID),
		"mid", utils.GobEncoderOutString(verifiedData.MID),
		"hsv", utils.GobEncoderOutString(verifiedData.HASHVEC),
		"ags", string(verifiedData.AGS),
		"res", utils.GobEncoderOutString(verifiedData.RES),
		"agd", string(verifiedData.AGGDATE),
		"rd", string(verifiedData.REGDATE),
	).Err()
	if err != nil {
		println("KVS HSet() Error:")
		fmt.Println(err)
	}
}

// GetVerifyData --
func (verifiedData *VerifyData) Get(ctx context.Context, db *redis.Client) {
	val, err := db.HGetAll(ctx, "vd:"+fmt.Sprint(verifiedData.AID)+":"+fmt.Sprint(verifiedData.VDID)).Result()
	if err != nil {
		println("KVS HGet() Error:")
		fmt.Println(err)
	}

	verifiedData.AID = []byte(val["aid"])
	verifiedData.VDID = utils.GetUint64FromString2Map("vdid", val)
	verifiedData.MID = utils.GobDecoderUint64Array(val["mid"])
	verifiedData.HASHVEC = utils.GobDecoderByteByteArray(val["hsv"])
	verifiedData.AGS = []byte(val["AGS"])
	verifiedData.RES = utils.StringConvertToBool(val["res"])
	verifiedData.AGGDATE = []byte(val["agd"])
	verifiedData.REGDATE = []byte(val["rd"])

	if err != nil {
		panic(err)
	}
}

// //SetFeedback --

func (feedback *Feedback) Set(ctx context.Context, db *redis.Client) {
	err := db.HSet(
		ctx,
		"fb:"+fmt.Sprint(utils.ByteArrayConvertToUint(feedback.AID))+":"+fmt.Sprint(utils.ByteArrayConvertToUint(feedback.VUID)),
		"aid:", string(feedback.AID),
		"sid:", string(feedback.SID),
		"vid:", string(feedback.VUID),
		"rd", string(feedback.REGDATE),
		"sig", string(feedback.VSIGN),
		"td", string(feedback.REGDATE),
	).Err()
	if err != nil {
		println("KVS HSet() Error:")
		fmt.Println(err)
	}
}

func (feedback *Feedback) SetTrackedData(ctx context.Context, db *redis.Client) {
	err := db.HSet(
		ctx,
		"fb:"+fmt.Sprint(utils.ByteArrayConvertToUint(feedback.AID))+":"+fmt.Sprint(utils.ByteArrayConvertToUint(feedback.VUID)),
		"td", string(feedback.TRACKEDDATE),
	).Err()
	if err != nil {
		println("KVS HSet() Error:")
		fmt.Println(err)
	}
}

// GetFeedback --
func (feedback *Feedback) Get(ctx context.Context, db *redis.Client) {

	val, err := db.HGetAll(ctx, "fb:"+fmt.Sprint(utils.ByteArrayConvertToUint(feedback.AID))+":"+fmt.Sprint(utils.ByteArrayConvertToUint(feedback.VUID))).Result()
	if err != nil {
		println("KVS HGet() Error:")
		fmt.Println(err)
	}
	// aggregator.IP = val["ip/"]
	feedback.AID = []byte(val["aid"])
	feedback.SID = []byte(val["sid"])
	feedback.VUID = []byte(val["vid"])
	feedback.REGDATE = []byte(val["rd"])
	feedback.VSIGN = []byte(val["sig"])
	feedback.TRACKEDDATE = []byte(val["td"])
}
