package simplelru

type LRUCache interface {
	Set(k, v interface{})

	Get(k interface{}) (v interface{}, ok bool)

	Contains(k interface{}) bool

	Peek(k interface{}) (v interface{}, ok bool)

	Remove(k interface{}) bool

	RemoveOldest() (k, v interface{}, ok bool)

	Len() int

	Keys() []interface{}

	Purge()

	Resize(int) int
}
