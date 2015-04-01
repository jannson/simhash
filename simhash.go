package simhash

//#include "inc/simtable.h"
//#cgo LDFLAGS: -L/usr/local/lib -lJudy -lsimhash
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

type LongVector struct {
	p unsafe.Pointer
}

type SimTable struct {
	p    unsafe.Pointer
	lock sync.RWMutex
	//lock sync.Mutex
}

func NewLongVector() LongVector {
	var ret LongVector
	ret.p = C.LongVectorInit()
	return ret
}

func (v LongVector) Release() {
	C.LongVectorRelease(v.p)
}

func (v LongVector) Reserve(size int) {
	C.LongVectorReserve(v.p, C.long(size))
}

func (v LongVector) Add(n uint64) {
	C.LongVectorAdd(v.p, C.ulong(n))
}

func (v LongVector) Set(i int64, n uint64) {
	C.LongVectorSet(v.p, C.int(i), C.ulong(n))
}

func (v LongVector) Get(i int64) uint64 {
	return uint64(C.LongVectorGet(v.p, C.int(i)))
}

func (v LongVector) ToArrayPtr() (*C.ulong, int) {
	var l int64
	p := C.LongVector2Array(v.p, (*C.long)(unsafe.Pointer(&l)))
	return p, int(l)
}

func (v LongVector) Len() int64 {
	return int64(C.LongVectorLen(v.p))
}

func NewSimTable(d int64, v LongVector) SimTable {
	var ret SimTable
	ret.p = C.SimTableInit(C.long(d), v.p)
	return ret
}

func (st SimTable) Release() {
	st.lock.Lock()
	defer st.lock.Unlock()
	C.SimTableRelease(st.p)
}

func (st SimTable) Find(h uint64) uint64 {
	st.lock.RLock()
	defer st.lock.RUnlock()
	return uint64(C.SimTableFind(st.p, C.ulong(h)))
}

func (st SimTable) FindM(h uint64, v LongVector) {
	st.lock.RLock()
	defer st.lock.RUnlock()
	C.SimTableFindm(st.p, C.ulong(h), v.p)
}

func (st SimTable) Insert(h uint64) {
	st.lock.Lock()
	defer st.lock.Unlock()
	C.SimTableInsert(st.p, C.ulong(h))
}

func (st SimTable) InsertBulk(hashs []uint64) {
	st.lock.Lock()
	defer st.lock.Unlock()
	C.SimTableInsertBulk(st.p, (*C.ulong)(unsafe.Pointer(&hashs[0])), C.long(len(hashs)))
}

func (st SimTable) Remove(h uint64) {
	st.lock.Lock()
	defer st.lock.Unlock()
	C.SimTableRemove(st.p, C.ulong(h))
}

func (st SimTable) Permute(h uint64) uint64 {
	st.lock.Lock()
	defer st.lock.Unlock()
	return uint64(C.SimTablePermute(st.p, C.ulong(h)))
}

func (st SimTable) Unpermute(h uint64) {
	st.lock.Lock()
	defer st.lock.Unlock()
	C.SimTableUnpermute(st.p, C.ulong(h))
}

func (st SimTable) SearchMask() uint64 {
	st.lock.RLock()
	defer st.lock.RUnlock()
	return uint64(C.SimTableSearchMask(st.p))
}

type Corpus struct {
	tables    []SimTable
	diff_bits int
}

func combinations(n, m int, f func([]int)) {
	// For each combination of m elements out of n
	// call the function f passing a list of m integers in 0-n
	// without repetitions

	// TODO: switch to iterative algo
	s := make([]int, m)
	last := m - 1
	var rc func(int, int)
	rc = func(i, next int) {
		for j := next; j < n; j++ {
			s[i] = j
			if i == last {
				f(s)
			} else {
				rc(i+1, j+1)
			}
		}
		return
	}
	rc(0, 0)
}

func NewCorpus(num_blocks int, diff_bits int) *Corpus {
	corpus := &Corpus{nil, diff_bits}
	perms := []uint64{}
	for i := 0; i < num_blocks; i++ {
		start := (i * 64) / num_blocks
		end := ((i + 1) * 64) / num_blocks
		num := uint64(0)
		for j := start; j < end; j++ {
			num |= (uint64(1) << uint64(j))
		}
		perms = append(perms, num)
	}

	corpus.tables = []SimTable{}
	combinations(len(perms), num_blocks-diff_bits, func(ret []int) {
		lv := NewLongVector()
		defer lv.Release()
		cset := make(map[uint64]bool)

		for _, x := range ret {
			px := perms[x]
			cset[px] = true
			lv.Add(px)
		}
		for _, x := range perms {
			if _, ok := cset[x]; !ok {
				cset[x] = true
				lv.Add(x)
			}
		}

		corpus.tables = append(corpus.tables, NewSimTable(int64(diff_bits), lv))
	})

	return corpus
}

func (corpus *Corpus) Release() {
	for _, tb := range corpus.tables {
		tb.Release()
	}
}

func (corpus *Corpus) Insert(hash uint64) {
	for _, tb := range corpus.tables {
		tb.Insert(hash)
	}
}

func (corpus *Corpus) InsertBulk(hashs []uint64) {
	for _, tb := range corpus.tables {
		go func(gtb SimTable) {
			gtb.InsertBulk(hashs)
		}(tb)
	}
}

func (corpus *Corpus) Remove(hash uint64) {
	for _, tb := range corpus.tables {
		tb.Remove(hash)
	}
}

func (corpus *Corpus) Find(hash uint64) uint64 {
	for _, tb := range corpus.tables {
		if result := tb.Find(hash); result != 0 {
			return result
		}
	}

	return uint64(0)
}

func (corpus *Corpus) FindAll(hash uint64, f func(uint64)) {
	//filter := bloom.New(10000000, 5)
	//filter := make(map[uint64]bool)
	for _, tb := range corpus.tables {
		lv := NewLongVector()
		defer lv.Release()

		//start := time.Now()
		tb.FindM(hash, lv)
		//elapsed := time.Since(start)
		//fmt.Printf("FindM: %v\n", elapsed)

		//start = time.Now()
		//trick hear
		p, l := lv.ToArrayPtr()
		pa := (*[1 << 30]C.ulong)(unsafe.Pointer(p))
		for i := 0; i < l; i++ {
			key := uint64(pa[i])
			/* kb := make([]byte, 8)
			binary.LittleEndian.PutUint64(kb, key)
			if !filter.Test(kb) {
				filter.Add(kb)
				f(key)
			} */
			/*if _, ok := filter[key]; !ok {
				filter[key] = true
				f(key)
			}*/
			f(key)
		}
		//elapsed = time.Since(start)
		//fmt.Printf("Looop: %v\n", elapsed)
	}
}

func (corpus *Corpus) FindParallel(hash uint64, f func(uint64)) {
	corpusLen := len(corpus.tables)
	ch := make(chan int, corpusLen)

	lvs := make([]LongVector, corpusLen)
	for i := 0; i < corpusLen; i++ {
		lvs[i] = NewLongVector()
	}
	defer func() {
		for _, lv := range lvs {
			lv.Release()
		}
		close(ch)
	}()

	for idx, tb := range corpus.tables {
		go func(gst SimTable, i int) {
			gst.FindM(hash, lvs[i])
			ch <- i
		}(tb, idx)
	}

	for i := 0; i < corpusLen; i++ {
		idx := <-ch
		p, l := lvs[idx].ToArrayPtr()
		pa := (*[1 << 30]C.ulong)(unsafe.Pointer(p))
		for j := 0; j < l; j++ {
			f(uint64(pa[j]))
		}
	}
}

func (corpus *Corpus) Distance(a uint64, b uint64) int {
	x := (a ^ b)
	ans := 0
	for x != 0 {
		ans += 1
		x &= x - 1
	}
	return ans
}

func mainTest() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	lv := NewLongVector()
	defer lv.Release()

	lv.Add(4)
	lv.Add(3)
	lv.Set(0, 5)
	fmt.Println(lv.Get(0))
	fmt.Println(lv.Len())
	//trick hear
	pa, l := lv.ToArrayPtr()
	npa := (*[1 << 30]C.ulong)(unsafe.Pointer(pa))
	fmt.Printf("ToArray:%v, %v, %v\n", *pa, npa[1], l)

	b := 15
	x := NewCorpus(b+1, b)
	defer x.Release()

	start := time.Now()

	begin := 800
	end := 1800000
	hashs := make([]uint64, end-begin)
	for j := 800; j < end; j++ {
		hashs[j-begin] = uint64(j)
	}
	x.InsertBulk(hashs)

	elapsed := time.Since(start)
	fmt.Printf("Insert: %v\n", elapsed)

	m1 := make(map[uint64]bool)
	m2 := make(map[uint64]bool)

	start = time.Now()

	z := uint64(9001)
	for j := begin; j < end; j++ {
		if x.Distance(z, (uint64(j))) <= b {
			m1[uint64(j)] = true
		}
	}

	elapsed = time.Since(start)
	fmt.Printf("ShowAll: %v\n", elapsed)

	start = time.Now()
	i := 0
	x.FindParallel(z, func(ret uint64) {
		m2[ret] = true
		i++
	})

	elapsed = time.Since(start)
	fmt.Printf("FindAll: %v\n", elapsed)

	fmt.Printf("Len1:%d Len2:%d len3:%d\n", len(m1), len(m2), i)
}
