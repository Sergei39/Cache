package code

type Cache interface {
	Put(key uint32, value string)
	Get(key uint32) (string, bool)
}
