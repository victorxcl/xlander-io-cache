package cache

import (
	"log"
	"runtime"
	"strconv"
	"testing"
	"time"
)

type Person struct {
	Name     string
	Age      int
	Location string
}

func (p *Person) CacheBytes() int {
	return 100
}

func printMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("Alloc = %v KB, TotalAlloc = %v KB, Sys = %v KB,Lookups = %v NumGC = %v\n", m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.Lookups, m.NumGC)
}

func Test_Cache_Simple(t *testing.T) {

	cache, err := New(nil)

	if nil != err {
		t.Fatalf("New cache instance failed! err=%v", err)
	}

	jack := &Person{"Jack", 18, "London"}
	cache.Set("a", jack, 5)
	cache.Set("b", jack, 5)

	if int32(2) != cache.TotalItems() {
		t.Fatalf("cache's total items count should be 2, but %d", cache.TotalItems())
	}

	{
		v, ttl := cache.Get("a")
		log.Println("get a")
		log.Println(v, ttl)

		if v != jack {
			t.Fatalf("get 'a' from cache should be %v, but %v", jack, v)
		}

		if int64(5) != ttl {
			t.Fatalf("the ttl of object which get from key 'a' should be 5, but %d", ttl)
		}
	}

	log.Println("delete a")
	cache.Delete("a")

	if int32(1) != cache.TotalItems() {
		t.Fatalf("count of cache's total items should be 1, but %d", cache.TotalItems())
	}

	{
		v, ttl := cache.Get("a")
		log.Println("get a")
		log.Println(v, ttl)

		if nil != v {
			t.Fatalf("get 'a' from cache should be nil, but %v", v)
		}

		if int64(0) != ttl {
			t.Fatalf("the ttl of object which is not exist should be 0, but %d", ttl)
		}
	}

	{
		v, ttl := cache.Get("b")
		log.Println("get b")
		log.Println(v, ttl)

		if v != jack {
			t.Fatalf("get 'a' from cache should be %v, but %v", jack, v)
		}

		if int64(5) != ttl {
			t.Fatalf("the ttl of object which get from key 'b' should be 5, but %d", ttl)
		}
	}

	log.Println("origin a")
	log.Println(*jack)

	log.Println("To simulate TTLs to timeout, waiting for 10 seconds ...")
	time.Sleep(10 * time.Second) // wait timeout for all ttls

	if int32(0) != cache.TotalItems() {
		t.Fatalf("count of cache's total items should be 0 now, but %d", cache.TotalItems())
	}

	{
		v, ttl := cache.Get("a")
		log.Println("get a")
		log.Println(v, ttl)

		if nil != v {
			t.Fatalf("get 'a' from cache should be nil, but %v", v)
		}

		if int64(0) != ttl {
			t.Fatalf("the ttl of object which is not exist should be 0, but %d", ttl)
		}
	}
	{
		v, ttl := cache.Get("b")
		log.Println("get b")
		log.Println(v, ttl)

		if nil != v {
			t.Fatalf("get 'a' from cache should be nil, but %v", v)
		}

		if int64(0) != ttl {
			t.Fatalf("the ttl of object which is not exist should be 0, but %d", ttl)
		}
	}
}

func Test_Cache_Expire(t *testing.T) {
	cache, err := New(nil)

	if nil != err {
		t.Fatalf("New cache instance failed! err=%v", err)
	}

	jack := &Person{"Jack", 18, "London"}

	cache.Set("1", jack, 5)
	cache.Set("2", jack, 18)
	cache.Set("3", jack, 23)
	cache.Set("4", jack, -100)
	cache.Set("5", jack, 3000000)
	cache.Set("6", jack, 35)

	if int32(5) != cache.TotalItems() {
		t.Fatalf("count of cache's total items should be 5, but %d", cache.TotalItems())
	}

	count := 0
	for {
		v1, ttl1 := cache.Get("1")
		log.Printf("1==>%v %v", v1, ttl1)
		v2, ttl2 := cache.Get("2")
		log.Printf("2==>%v %v", v2, ttl2)
		v3, ttl3 := cache.Get("3")
		log.Printf("3==>%v %v", v3, ttl3)
		v4, ttl4 := cache.Get("4")
		log.Printf("4==>%v %v", v4, ttl4)
		v5, ttl5 := cache.Get("5")
		log.Printf("5==>%v %v", v5, ttl5)
		v6, ttl6 := cache.Get("6")
		log.Printf("6==>%v %v", v6, ttl6)
		log.Println("total key", cache.TotalItems())
		log.Println("-----------")

		if int(0) == count {
			if int32(5) != cache.TotalItems() {
				t.Fatalf("count of cache's total items should be 5, but %d", cache.TotalItems())
			}

			if v1 != jack {
				t.Fatalf("get '1' from cache should be %v, but %v", jack, v1)
			}
			if v2 != jack {
				t.Fatalf("get '2' from cache should be %v, but %v", jack, v2)
			}
			if v3 != jack {
				t.Fatalf("get '3' from cache should be %v, but %v", jack, v3)
			}
			if v4 != nil {
				t.Fatalf("get '4' from cache should be nil, but %v", v4)
			}
			if v5 != jack {
				t.Fatalf("get '5' from cache should be %v, but %v", jack, v5)
			}
			if v6 != jack {
				t.Fatalf("get '6' from cache should be %v, but %v", jack, v6)
			}

			if int64(5) != ttl1 {
				t.Fatalf("the ttl of object which get from key '1' should be 5, but %d", ttl1)
			}
			if int64(18) != ttl2 {
				t.Fatalf("the ttl of object which get from key '2' should be 18, but %d", ttl2)
			}
			if int64(23) != ttl3 {
				t.Fatalf("the ttl of object which get from key '3' should be 23, but %d", ttl3)
			}
			if int64(0) != ttl4 {
				t.Fatalf("the ttl of object which get from key '4' should be 0, but %d", ttl4)
			}
			if int64(7200) != ttl5 {
				t.Fatalf("the ttl of object which get from key '5' should be 7200, but %d", ttl5)
			}
			if int64(35) != ttl6 {
				t.Fatalf("the ttl of object which get from key '6' should be 35, but %d", ttl6)
			}
		}

		if int(10) == count {
			if int32(4) != cache.TotalItems() {
				t.Fatalf("count of cache's total items should be 4, but %d", cache.TotalItems())
			}

			if v1 != nil {
				t.Fatalf("get '1' from cache should be nil, but %v", v1)
			}
			if v2 != jack {
				t.Fatalf("get '2' from cache should be %v, but %v", jack, v2)
			}
			if v3 != jack {
				t.Fatalf("get '3' from cache should be %v, but %v", jack, v3)
			}
			if v4 != nil {
				t.Fatalf("get '4' from cache should be nil, but %v", v4)
			}
			if v5 != jack {
				t.Fatalf("get '5' from cache should be %v, but %v", jack, v5)
			}
			if v6 != jack {
				t.Fatalf("get '6' from cache should be %v, but %v", jack, v6)
			}

			if int64(0) != ttl1 {
				t.Fatalf("the ttl of object which get from key '1' should be 0, but %d", ttl1)
			}
			if int64(18-10) != ttl2 {
				t.Fatalf("the ttl of object which get from key '2' should be 18-10, but %d", ttl2)
			}
			if int64(23-10) != ttl3 {
				t.Fatalf("the ttl of object which get from key '3' should be 23-10, but %d", ttl3)
			}
			if int64(0) != ttl4 {
				t.Fatalf("the ttl of object which get from key '4' should be 0, but %d", ttl4)
			}
			if int64(7200-10) != ttl5 {
				t.Fatalf("the ttl of object which get from key '5' should be 7200-10, but %d", ttl5)
			}
			if int64(35-10) != ttl6 {
				t.Fatalf("the ttl of object which get from key '6' should be 35-10, but %d", ttl6)
			}
		}

		if int(20) == count {
			if int32(3) != cache.TotalItems() {
				t.Fatalf("count of cache's total items should be 3, but %d", cache.TotalItems())
			}

			if v1 != nil {
				t.Fatalf("get '1' from cache should be nil, but %v", v1)
			}
			if v2 != nil {
				t.Fatalf("get '2' from cache should be nil, but %v", v2)
			}
			if v3 != jack {
				t.Fatalf("get '3' from cache should be %v, but %v", jack, v3)
			}
			if v4 != nil {
				t.Fatalf("get '4' from cache should be nil, but %v", v4)
			}
			if v5 != jack {
				t.Fatalf("get '5' from cache should be %v, but %v", jack, v5)
			}
			if v6 != jack {
				t.Fatalf("get '6' from cache should be %v, but %v", jack, v6)
			}

			if int64(0) != ttl1 {
				t.Fatalf("the ttl of object which get from key '1' should be 0, but %d", ttl1)
			}
			if int64(0) != ttl2 {
				t.Fatalf("the ttl of object which get from key '2' should be 0, but %d", ttl2)
			}
			if int64(23-20) != ttl3 {
				t.Fatalf("the ttl of object which get from key '3' should be 23-20, but %d", ttl3)
			}
			if int64(0) != ttl4 {
				t.Fatalf("the ttl of object which get from key '4' should be 0, but %d", ttl4)
			}
			if int64(7200-20) != ttl5 {
				t.Fatalf("the ttl of object which get from key '5' should be 7200-20, but %d", ttl5)
			}
			if int64(35-20) != ttl6 {
				t.Fatalf("the ttl of object which get from key '6' should be 35-20, but %d", ttl6)
			}
		}

		if int(30) == count {
			if int32(2) != cache.TotalItems() {
				t.Fatalf("count of cache's total items should be 2, but %d", cache.TotalItems())
			}

			if v1 != nil {
				t.Fatalf("get '1' from cache should be nil, but %v", v1)
			}
			if v2 != nil {
				t.Fatalf("get '2' from cache should be nil, but %v", v2)
			}
			if v3 != nil {
				t.Fatalf("get '3' from cache should be nil, but %v", v3)
			}
			if v4 != nil {
				t.Fatalf("get '4' from cache should be nil, but %v", v4)
			}
			if v5 != jack {
				t.Fatalf("get '5' from cache should be %v, but %v", jack, v5)
			}
			if v6 != jack {
				t.Fatalf("get '6' from cache should be %v, but %v", jack, v6)
			}

			if int64(0) != ttl1 {
				t.Fatalf("the ttl of object which get from key '1' should be 0, but %d", ttl1)
			}
			if int64(0) != ttl2 {
				t.Fatalf("the ttl of object which get from key '2' should be 0, but %d", ttl2)
			}
			if int64(0) != ttl3 {
				t.Fatalf("the ttl of object which get from key '3' should be 0, but %d", ttl3)
			}
			if int64(0) != ttl4 {
				t.Fatalf("the ttl of object which get from key '4' should be 0, but %d", ttl4)
			}
			if int64(7200-30) != ttl5 {
				t.Fatalf("the ttl of object which get from key '5' should be 7200-30, but %d", ttl5)
			}
			if int64(35-30) != ttl6 {
				t.Fatalf("the ttl of object which get from key '6' should be 35-30, but %d", ttl6)
			}
		}

		if int(40) == count {
			if int32(1) != cache.TotalItems() {
				t.Fatalf("count of cache's total items should be 2, but %d", cache.TotalItems())
			}

			if v1 != nil {
				t.Fatalf("get '1' from cache should be nil, but %v", v1)
			}
			if v2 != nil {
				t.Fatalf("get '2' from cache should be nil, but %v", v2)
			}
			if v3 != nil {
				t.Fatalf("get '3' from cache should be nil, but %v", v3)
			}
			if v4 != nil {
				t.Fatalf("get '4' from cache should be nil, but %v", v4)
			}
			if v5 != jack {
				t.Fatalf("get '5' from cache should be %v, but %v", jack, v5)
			}
			if v6 != nil {
				t.Fatalf("get '6' from cache should be nil, but %v", v6)
			}

			if int64(0) != ttl1 {
				t.Fatalf("the ttl of object which get from key '1' should be 0, but %d", ttl1)
			}
			if int64(0) != ttl2 {
				t.Fatalf("the ttl of object which get from key '2' should be 0, but %d", ttl2)
			}
			if int64(0) != ttl3 {
				t.Fatalf("the ttl of object which get from key '3' should be 0, but %d", ttl3)
			}
			if int64(0) != ttl4 {
				t.Fatalf("the ttl of object which get from key '4' should be 0, but %d", ttl4)
			}
			if int64(7200-40) != ttl5 {
				t.Fatalf("the ttl of object which get from key '5' should be 7200-40, but %d", ttl5)
			}
			if int64(0) != ttl6 {
				t.Fatalf("the ttl of object which get from key '6' should be 0, but %d", ttl6)
			}
		}

		count++
		if count > 45 {
			return
		}
		time.Sleep(time.Second)
	}
}

// func Test_SetAndRemove(t *testing.T) {
// 	a := Person{"Jack", 18, "America"}
// 	lc := New()

// 	log.Println("start")
// 	printMemStats()

// 	for i := 0; i < 20; i++ {
// 		//set
// 		for j := 0; j < 10000; j++ {
// 			lc.Set(strconv.Itoa(j), a, 1)
// 		}

// 		log.Println("round:", i)
// 		log.Println("finish set")
// 		printMemStats()

// 		time.Sleep(2 * time.Second)
// 	}

// 	log.Println("finish")
// 	printMemStats()
// }

// func Test_BigAmountKey(t *testing.T) {
// 	a := Person{"Jack", 18, "America"}
// 	lc := New()

// 	log.Println("start")
// 	printMemStats()

// 	go func() {
// 		log.Println(http.ListenAndServe("0.0.0.0:10000", nil))
// 	}()

// 	for i := 0; i < 30; i++ {
// 		log.Println("----------")
// 		log.Println("round", i)
// 		log.Println("mem start set")
// 		printMemStats()

// 		for i := 0; i < 1000000; i++ {
// 			lc.Set(strconv.Itoa(i), a, int64(rand.Intn(10)+1))
// 		}
// 		log.Println("mem after set")
// 		printMemStats()
// 		time.Sleep(time.Second)
// 	}

// 	log.Println("~~~~~~")
// 	log.Println("finish set")
// 	printMemStats()

// 	log.Println("do GC")

// 	runtime.GC()
// 	log.Println("after GC")
// 	printMemStats()

// 	count := 0
// 	for {
// 		time.Sleep(1 * time.Second)
// 		log.Println("---job finished---")
// 		printMemStats()
// 		count++
// 		if count > 45 {
// 			//return
// 		}
// 	}
// 	//time.Sleep(1*time.Hour)
// }

// func Test_RandSet(t *testing.T) {
// 	lc, _ := New(nil)
// 	a := Person{"Jack", 18, "America"}

// 	lc.Set("a", a, 15)
// 	lc.Set("b", a, 19)
// 	lc.Set("c", a, 60)
// 	lc.Set("d", a, 63)
// 	lc.Set("e", a, 65)

// 	log.Println("before big amount set")
// 	v, ttl := lc.Get("a")
// 	log.Printf("a==>%v %v", v, ttl)
// 	v, ttl = lc.Get("b")
// 	log.Printf("b==>%v %v", v, ttl)
// 	v, ttl = lc.Get("c")
// 	log.Printf("c==>%v %v", v, ttl)
// 	v, ttl = lc.Get("d")
// 	log.Printf("d==>%v %v", v, ttl)
// 	v, ttl = lc.Get("e")
// 	log.Printf("e==>%v %v", v, ttl)

// 	log.Println("start amount set")
// 	for i := 0; i < 200; i++ {
// 		for j := 0; j < 10000; j++ {
// 			num := rand.Intn(9999999999999)
// 			key := strconv.Itoa(num)
// 			lc.Set(key, a, int64(rand.Intn(30)+20))
// 		}
// 	}

// 	for i := 0; i < 70; i++ {
// 		time.Sleep(time.Second)
// 		log.Println("--------------")
// 		v, ttl = lc.Get("a")
// 		log.Printf("a==>%v %v", v, ttl)
// 		v, ttl = lc.Get("b")
// 		log.Printf("b==>%v %v", v, ttl)
// 		v, ttl = lc.Get("c")
// 		log.Printf("c==>%v %v", v, ttl)
// 		v, ttl = lc.Get("d")
// 		log.Printf("d==>%v %v", v, ttl)
// 		v, ttl = lc.Get("e")
// 		log.Printf("e==>%v %v", v, ttl)
// 		log.Println("total key", lc.GetLen())
// 	}
// }

// func Test_KeepTTL(t *testing.T) {
// 	lc := New()
// 	a := Person{"Ma Yun", 58, "China"}
// 	b := Person{"Jack Ma", 18, "America"}

// 	lc.Set("a", a, 30)
// 	lc.Set("b", a, 40)
// 	lc.Set("c", a, 50)

// 	//log
// 	v, ttl := lc.Get("a")
// 	log.Printf("a==>%v %v", v, ttl)
// 	v, ttl = lc.Get("b")
// 	log.Printf("b==>%v %v", v, ttl)
// 	v, ttl = lc.Get("c")
// 	log.Printf("c==>%v %v", v, ttl)

// 	time.Sleep(5 * time.Second)

// 	lc.Set("a", b, 300)
// 	lc.Set("b", b, 0)

// 	//log
// 	for i := 0; i < 10; i++ {
// 		log.Println("-----------")
// 		v, ttl = lc.Get("a")
// 		log.Printf("a==>%v %v", v, ttl)
// 		v, ttl = lc.Get("b")
// 		log.Printf("b==>%v %v", v, ttl)
// 		v, ttl = lc.Get("c")
// 		log.Printf("c==>%v %v", v, ttl)
// 		time.Sleep(time.Second)
// 	}

// }

// func Test_SetTTL(t *testing.T) {
// 	lc := New()
// 	a := Person{"Ma Yun", 58, "China"}

// 	ttls := []int64{1, 20000, 0, -100, 200, 45, 346547457457457, -20000, 434, 9}
// 	for i := 0; i < 10; i++ {
// 		key := strconv.Itoa(i)
// 		lc.Set(key, a, ttls[i])
// 	}

// 	for i := 0; i < 10; i++ {
// 		log.Println("-----------")
// 		for j := 0; j < 10; j++ {
// 			key := strconv.Itoa(j)
// 			v, ttl := lc.Get(key)
// 			log.Printf("%s==>%v %v", key, v, ttl)
// 		}
// 		log.Println("total key", lc.GetLen())
// 		time.Sleep(time.Second)
// 	}
// }

// func Test_SyncMap(t *testing.T) {
// 	printMemStats()

// 	go func() {
// 		log.Println(http.ListenAndServe("0.0.0.0:10000", nil))
// 	}()

// 	type Person struct {
// 		Name     string
// 		Age      int
// 		Location string
// 	}
// 	a := Person{"Jack", 18, "America"}
// 	type Element struct {
// 		Member string
// 		Score  int64
// 		Value  interface{}
// 	}

// 	var myMap sync.Map
// 	for i := 0; i < 1000000; i++ {
// 		key := strconv.Itoa(i)
// 		b := &Element{
// 			Member: key,
// 			Score:  10,
// 			Value:  a,
// 		}
// 		myMap.Store(key, b)
// 	}

// 	printMemStats()

// 	for i := 0; i < 1000000; i++ {
// 		myMap.Delete(strconv.Itoa(i))
// 	}
// 	runtime.GC()

// 	for {
// 		printMemStats()
// 		time.Sleep(time.Second)
// 	}
// }

func BenchmarkLocalReference_SetPointer(b *testing.B) {
	lc, _ := New(nil)
	a := &Person{"Jack", 18, "America"}

	keyArray := []string{}
	for i := 0; i < b.N; i++ {
		keyArray = append(keyArray, strconv.Itoa(i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Set(keyArray[i], a, 300)
	}
}

func BenchmarkLocalReference_GetPointer(b *testing.B) {
	lc, _ := New(nil)
	a := &Person{"Jack", 18, "America"}
	lc.Set("1", a, 300)
	var e *Person
	log.Println(e)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it, _ := lc.Get("1")
		e = it.(*Person)
	}
}

// func Benchmark_syncMap(b *testing.B) {
// 	var m sync.Map
// 	a := &Person{"Jack", 18, "America"}
// 	for i := 0; i < 100; i++ {
// 		m.Store(i, a)
// 	}

// 	b.ReportAllocs()
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		p, _ := m.Load(1)
// 		b := &Person{"Jack", 18, "America"}
// 		m.Store(i, b)
// 		_ = p.(*Person)
// 	}

// }

// func Benchmark_map(b *testing.B) {
// 	m := map[int]int{}
// 	for i := 0; i < 100; i++ {
// 		m[i] = i
// 	}

// 	b.ReportAllocs()
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = m[i]
// 	}

// }

// func Benchmark_time(b *testing.B) {

// 	b.ReportAllocs()
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		time.Now().Unix()
// 	}

// }
