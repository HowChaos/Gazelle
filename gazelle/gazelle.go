package gazelle

import (
	"fmt"
	"log"
	"sync"
)

//identify by name
type Group struct {
	name   string //identify
	getter Getter //callback method
	Cache  cache  //the cache
}

var (
	mutex sync.RWMutex
	group = make(map[string]*Group) //store all groups
)

//a interface function, achieve a interface with call itself
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

//initialization a new group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mutex.Lock()
	defer mutex.Unlock()
	g := &Group{
		name:   name,
		getter: getter,
		Cache: cache{
			cacheBytes: cacheBytes,
		},
	}
	group[name] = g
	return g
}

func GetGroup(name string) *Group {
	mutex.RLock()
	g := group[name]
	mutex.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.Cache.Get(key); ok {
		log.Println("[Gazelle] hit")
		return v, nil
	}
	return g.Load(key)
}

func (g *Group) Load(key string) (value ByteView, err error) {
	return g.GetLocally(key)
}

func (g *Group) GetLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{cloneByte(bytes)}
	g.PopulateCache(key, value)
	return value, nil
}

func (g *Group) PopulateCache(key string, value ByteView) {
	g.Cache.Add(key, value)
}
