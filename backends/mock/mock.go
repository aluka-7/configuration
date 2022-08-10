package mock

func NewMockClient(store map[string]string) (*Client, error) {
	return &Client{store: store}, nil
}

type Client struct {
	store map[string]string
}

func (c Client) Add(path string, value []byte) (string, error) {
	panic("implement me")
}

func (c Client) GetValues(keys []string) (vls map[string]string, err error) {
	vls = make(map[string]string, len(keys))
	for _, v := range keys {
		vls[v] = c.store[v]
	}
	return
}
func (c Client) WatchPrefix(keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	return 0, nil
}
