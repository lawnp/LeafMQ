package nixmq

type Users struct {
	Users map[string]string
}

func NewUsers() *Users {
	return &Users{
		Users: make(map[string]string),
	}
}

// testing purposes only
func (u *Users) PopulateUsers() {
	u.Add("admin", "admin")
	u.Add("user", "user")
}

func (u *Users) Add(username, password string) {
	u.Users[username] = password
}

func (u *Users) Get(username string) (string, bool) {
	password, ok := u.Users[username]
	return password, ok
}

func (u *Users) Remove(username string) {
	delete(u.Users, username)
}
