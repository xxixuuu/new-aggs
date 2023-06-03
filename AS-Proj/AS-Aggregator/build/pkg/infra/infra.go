package infra

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/herumi/bls-go-binary/bls"
)

type SignerBuffer struct {
	Buf [][]byte
}

type AggregatorSignerEachBuffer struct {
	SignerBufArray []*SignerBuffer
}

type AggregatorBuffer struct {
	Buf [][]byte //allin one
}

//
type MessageData struct {
	SID                []byte //4byte
	MessageHeader      []byte //4byte
	Message            []byte
	SignerOptionHeader []byte
	SignerOption       []byte
	Sign               bls.Sign
}

type Child struct {
	ChPtr   *chan *MessageData
	TcpConn *net.TCPConn
	Sid     uint32
}

func NewChild(chprt chan *MessageData, conn *net.TCPConn, sid uint32) *Child {
	return &Child{
		ChPtr: &chprt, TcpConn: conn, Sid: sid,
	}
}

//
type ChildMap struct {
	dictionary map[uint32]*Child //if nil means child isDead
	sync.Mutex
}

//
func NewChildMap() *ChildMap {
	return &ChildMap{
		dictionary: make(map[uint32]*Child),
	}
}

//
func (m *ChildMap) Set(SID uint32, child *Child) {
	// println("s- Waiting For UnLock")
	m.Lock()
	// println("s- Lock")
	m.dictionary[SID] = child
	// println("s- UnLock")
	m.Unlock()
}

//
func (m *ChildMap) DeleteKey(SID uint32) {
	m.Lock()
	delete(m.dictionary, SID)
	m.Unlock()
}

//
func (m *ChildMap) Get(SID uint32) *Child { //0 = isDead ,1 = isAlive, 2 = DosentExist
	// println("g- aiting For UnLock")
	m.Lock()
	// println("g- Lock")
	res := m.dictionary[SID]
	// println("g- UnLock")
	m.Unlock()
	return res
}

func (m *ChildMap) GetDistinctMessage(SID uint32) *MessageData { //0 = isDead ,1 = isAlive, 2 = DosentExist
	var res *Child
	m.Lock()
	res = m.dictionary[SID]
	m.Unlock()
	return <-*res.ChPtr
}

func (m *ChildMap) GetLastAliveKey() uint32 {
	m.Lock()
	keys := make([]int, len(m.dictionary))

	i := 0
	for k := range m.dictionary {
		keys[i] = int(k)
		i++
	}
	m.Unlock()
	if len(keys) > 0 {
		sort.Ints(keys)
		// fmt.Println(keys)
		return uint32(keys[len(keys)-1])
	} else {
		return 0
	}
}

func (m *ChildMap) GetRandomMessage() *MessageData { //0 = isDead ,1 = isAlive, 2 = DosentExist

	var res *Child
	for {
		lastAliveKey := m.GetLastAliveKey()

		if lastAliveKey > 0 {
			target := rand.Uint32() % (lastAliveKey + 1)
			// fmt.Println("target", target)
			if m.GetAlive(target) > 0 {
				res = m.Get(target)
				if len(*res.ChPtr) > 0 {
					return <-*res.ChPtr
				}
			}
		} else {
			time.Sleep(time.Second * 3)
			fmt.Println("No more signer")
		}
	}
}

func (m *ChildMap) GetAlive(SID uint32) int { //if connection closed return 0, alive return Sid, not exist -1
	m.Lock()
	res, ok := m.dictionary[SID]

	if ok {
		if res.ChPtr != nil {
			m.Unlock()
			return int(res.Sid)
		} else {
			m.Unlock()
			return 0
		}

	}
	m.Unlock()
	return -1
}

//
func (m *ChildMap) CheckAlivesAhead(index uint32) int {
	var aliveNum int = 0
	index++
	for {
		res := m.GetAlive(index)
		if res > 0 {
			aliveNum++
		} else {
			break
		}
		index++
	}
	return aliveNum
}
func (m *ChildMap) GetAlivesAhead(lastIndex uint32, nums int) []uint32 {
	index := lastIndex
	var indexArray []uint32
	for {
		res := m.GetAlive(uint32(index))
		if res > 0 {
			indexArray = append(indexArray, index)
			nums--
		}
		if nums < 1 {
			break
		}
		index++
	}
	return indexArray
}
func (m *ChildMap) CountAlive() int {
	sid := 1
	alivingSigners := 0
	//count aliving signers
	for {
		mapRes := m.GetAlive(uint32(sid))
		if mapRes >= 0 {
			if mapRes > 0 {
				alivingSigners++
			}
		} else {
			break
		}
		sid++
	}
	return alivingSigners
}

func (m *ChildMap) CountAll() int {
	sid := 1
	alivingSigners := 0
	//count aliving signers
	for {
		mapRes := m.GetAlive(uint32(sid))
		if mapRes <= 0 {
			break
		}
		alivingSigners++
		sid++
	}
	return alivingSigners
}

func (m *ChildMap) GetSignersFrom(num int, index int) []int { //index | index+1, index+2 ....
	var res []int = make([]int, 0)
	count := 1
	for {
		if m.GetAlive(uint32(index+count)) > 0 {
			res = append(res, index+count)
		}
		if len(res) >= num {
			break
		}
		count++
	}

	return res

}

func (m *ChildMap) ReverseWalk(index int, steps int) int {
	for {
		if m.GetAlive(uint32(index)) > 0 {
			steps--
		}
		if steps <= 0 {
			break
		}

		if index < 0 {
			fmt.Println("backwalk failed, index out of range")
			fmt.Println("index:", index)
			fmt.Println("steps:", steps)
			break
		}
		index--

	}
	return index
}

func CreateChildMap(SM *ChildMap) []byte {
	var i uint32 = 0
	var data []byte
	var netData []byte

	for {
		i++
		res := SM.GetAlive(i)

		if res < 0 {
			netData = append(netData, []byte("M")...)
			//here
			var buff []byte = make([]byte, 4)
			binary.LittleEndian.PutUint32(buff, i-1)
			netData = append(netData, buff[:]...)
			netData = append(netData, data[:]...)
			break
		} else {
			var buf []byte = make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, i)
			data = append(data, buf[:]...)
			data = append(data, uint8(res))
		}
	}
	return netData

}

//For multiChildHandler Load management

type MultiHandlerMap struct {
	sync.Mutex
	dictionary map[int]int //[HandlerID]:Holding Number
}

func NewMultiHandlerMap() *MultiHandlerMap {
	return &MultiHandlerMap{
		dictionary: make(map[int]int, 16384),
	}
}

func (m *MultiHandlerMap) Set(HID int, holds int) {
	m.Lock()
	m.dictionary[HID] = holds
	m.Unlock()
}

//
func (m *MultiHandlerMap) Get(HID int) int { //0 = isDead ,1 = isAlive, 2 = DosentExist

	m.Lock()
	if val, ok := m.dictionary[HID]; ok {
		m.Unlock()
		return val
	}
	m.Unlock()
	return -1
}

func (m *MultiHandlerMap) Inc(HID int) { //0 = isDead ,1 = isAlive, 2 = DosentExist
	m.Lock()
	m.dictionary[HID]++
	m.Unlock()
}

func (m *MultiHandlerMap) Dec(HID int) { //0 = isDead ,1 = isAlive, 2 = DosentExist
	m.Lock()
	m.dictionary[HID]--
	m.Unlock()
}

func (m *MultiHandlerMap) CheckLoadbalance(HID int) bool { //false = do noting, true = take task
	i := 0
	selfLoad := m.Get(HID)
	LoadSlice := make([]int, 0)
	for {
		if i != HID {
			load := m.Get(i)
			if load < 0 {
				break
			}
			LoadSlice = append(LoadSlice, load)
		}
		i++
	}
	sort.Slice(LoadSlice, func(i2, j int) bool {
		return LoadSlice[i2] < LoadSlice[j]
	})

	if len(LoadSlice) <= 0 || selfLoad <= LoadSlice[0] {
		return true
	}
	return false
}
