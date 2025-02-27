package repo

type Repo struct {
	values map[string]string
}

var instance *Repo

func Get() *Repo {
	return instance
}

func init() {
	instance = NewRepo()
}

func NewRepo() *Repo {
	var c Repo
	c.values = make(map[string]string)
	return &c
}

func (c *Repo) Add(key, value string) {
	c.values[key] = value
}

func (c *Repo) Get(key string) string {
	return c.values[key]
}
