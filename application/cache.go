package application

// AccessoryCache should serve to store accessories and associate metadata - tags (like room, etc)
type AccessoryCache struct {
	path string
}

func NewAccessoryCache(dir string) *AccessoryCache {
	return &AccessoryCache{
		path: dir,
	}
}

func (c *AccessoryCache) Save(deviceId string, aid uint64, tags map[string]struct{}) error {
	return nil
}

func (c *AccessoryCache) GetTags(deviceId string, aid uint64) map[string]struct{} {
	return nil
}
